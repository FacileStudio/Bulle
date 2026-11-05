import {
	claimDevice,
	relayUrl,
	revokeDevice,
	type ClaimedDevice,
	type HardwareDevice
} from './api';
import { OPCODE_AUDIO, OPCODE_CONTROL, createReassembler } from './frames';
import {
	createFrameWriter,
	isSecurityError,
	listen,
	openRpcChannel,
	type Characteristic,
	type FrameWriter,
	type RpcChannel
} from './gatt';

export const BOHS_SERVICE = 'b0010000-f4c1-4e00-b001-000000000001';
const INFO_CHAR = 'b0010001-f4c1-4e00-b001-000000000001';
const AUDIO_CHAR = 'b0010002-f4c1-4e00-b001-000000000001';
const CONTROL_CHAR = 'b0010003-f4c1-4e00-b001-000000000001';
const PROV_CHAR = 'b0010004-f4c1-4e00-b001-000000000001';

const AUTH_WINDOW_MS = 60_000;
const STATUS_POLL_MS = 1500;
const STATUS_REVOKED = 4401;
const RECONNECT_MIN_MS = 1000;
const RECONNECT_MAX_MS = 30_000;
const WRITE_FAILURES_REPORTED = 3;
const HARDWARE_ID = /^[0-9a-f]{8}$/;

type GattServer = {
	getPrimaryService(uuid: string): Promise<{ getCharacteristic(uuid: string): Promise<Characteristic> }>;
};

type BluetoothDevice = {
	id: string;
	name?: string;
	gatt?: { connected: boolean; connect(): Promise<GattServer>; disconnect(): void };
	addEventListener(type: 'gattserverdisconnected', listener: () => void): void;
};

type Bluetooth = {
	requestDevice(options: { filters: { services: string[] }[] }): Promise<BluetoothDevice>;
};

type Characteristics = Record<'info' | 'audio' | 'control' | 'prov', Characteristic>;

export type DeviceInfo = {
	version: number;
	state: number;
	capabilities: number;
	battery: number;
	firmware: string;
	transport: number;
};

/** The state of one device's relay: `live` turns false once it stops for good. */
export type RelayState = { live: boolean; message: string };

export type RelayOptions = {
	onStatus?: (message: string) => void;
	onRelay?: (deviceId: string, state: RelayState) => void;
};

export type PairOptions = RelayOptions & {
	name?: string;
	wifiSsid?: string;
	wifiPsk?: string;
};

export type PairedDevice = {
	device: HardwareDevice;
	info: DeviceInfo;
	close(): void;
};

/** Reports whether this browser exposes Web Bluetooth. */
export function bluetoothAvailable(): boolean {
	return typeof navigator !== 'undefined' && 'bluetooth' in navigator;
}

/**
 * Pairs a Pastille over Web Bluetooth and starts relaying it: asks the device
 * to identify, waits for the user to tap it, claims it on the account, writes
 * the gateway URL, token and optional Wi-Fi credentials, then forwards audio and
 * control frames between the device and /api/hardware/ws until `close()`.
 */
export async function scanAndPairBohsDevice(options: PairOptions = {}): Promise<PairedDevice> {
	const status = options.onStatus ?? (() => undefined);
	const ble = await chooseDevice();
	try {
		status('Connecting to your Pastille');
		const chars = await connect(ble);
		const info = parseInfo(await chars.info.readValue());
		const write = createFrameWriter();
		const prov = await openRpcChannel(chars.prov, write, () => ble.gatt?.disconnect());
		const identified = await prov.call('prov.identify');
		status('Tap your Pastille to confirm pairing');
		const authorized = await waitForAuthorization(prov);
		const hardwareId = reportedHardwareId(authorized) ?? requireHardwareId(identified);
		status('Adding it to your account');
		const claimed = await claimDevice(hardwareId, options.name || ble.name);
		await configure(prov, claimed, options);
		const relay = startRelay(ble, chars, write, claimed.device.id, options);
		status('Paired');
		return { device: claimed.device, info, close: relay.close };
	} catch (error) {
		ble.gatt?.disconnect();
		throw readable(error);
	}
}

/**
 * Relays a device already paired to the account, without pairing it again:
 * connects over Web Bluetooth, checks the device reports the expected hardware
 * id, then forwards frames until `close()`.
 */
export async function connectBohsDevice(device: HardwareDevice, options: RelayOptions = {}): Promise<PairedDevice> {
	const status = options.onStatus ?? (() => undefined);
	const ble = await chooseDevice();
	try {
		status(`Connecting to ${device.name}`);
		const chars = await connect(ble);
		const info = parseInfo(await chars.info.readValue());
		const write = createFrameWriter();
		const prov = await openRpcChannel(chars.prov, write, () => ble.gatt?.disconnect());
		if (requireHardwareId(await prov.call('prov.status')) !== device.hardware_id) {
			throw new Error(`The Pastille you picked is not ${device.name}.`);
		}
		const relay = startRelay(ble, chars, write, device.id, options);
		return { device, info, close: relay.close };
	} catch (error) {
		ble.gatt?.disconnect();
		throw readable(error);
	}
}

async function chooseDevice(): Promise<BluetoothDevice> {
	if (!bluetoothAvailable()) throw new Error('This browser has no Web Bluetooth. Use Chrome or Edge.');
	const bluetooth = (navigator as unknown as { bluetooth: Bluetooth }).bluetooth;
	try {
		return await bluetooth.requestDevice({ filters: [{ services: [BOHS_SERVICE] }] });
	} catch (error) {
		if (error instanceof Error && error.name === 'NotFoundError') throw new Error('No Pastille was selected.');
		throw readable(error);
	}
}

async function connect(ble: BluetoothDevice): Promise<Characteristics> {
	if (!ble.gatt) throw new Error('This Pastille has no GATT server.');
	const service = await (await ble.gatt.connect()).getPrimaryService(BOHS_SERVICE);
	const [info, audio, control, prov] = await Promise.all(
		[INFO_CHAR, AUDIO_CHAR, CONTROL_CHAR, PROV_CHAR].map((uuid) => service.getCharacteristic(uuid))
	);
	return { info, audio, control, prov };
}

function parseInfo(view: DataView): DeviceInfo {
	if (view.byteLength < 8) throw new Error('Pastille sent a device info record shorter than 8 bytes.');
	const byte = (at: number) => view.getUint8(at);
	return {
		version: byte(0),
		state: byte(1),
		capabilities: byte(2),
		battery: byte(3),
		firmware: `${byte(4)}.${byte(5)}.${byte(6)}`,
		transport: byte(7)
	};
}

function waitForAuthorization(prov: RpcChannel): Promise<unknown> {
	let stop = () => {};
	let poll: ReturnType<typeof setInterval> | undefined;
	let timeout: ReturnType<typeof setTimeout> | undefined;
	return new Promise<unknown>((resolve, reject) => {
		stop = prov.onState((state, message) => {
			if (state === 'authorized') resolve(message.result ?? message.params);
		});
		poll = setInterval(() => void prov.call('prov.status').catch(() => undefined), STATUS_POLL_MS);
		timeout = setTimeout(
			() => reject(new Error('Pastille was not tapped within 60 seconds.')),
			AUTH_WINDOW_MS
		);
	}).finally(() => {
		stop();
		clearInterval(poll);
		clearTimeout(timeout);
	});
}

/** Answers the 8 lowercase hex hardware id of a prov status result, or undefined. */
export function reportedHardwareId(status: unknown): string | undefined {
	const reported = (status as { hardware_id?: unknown } | undefined)?.hardware_id;
	return typeof reported === 'string' && HARDWARE_ID.test(reported) ? reported : undefined;
}

function requireHardwareId(status: unknown): string {
	const hardwareId = reportedHardwareId(status);
	if (!hardwareId) {
		throw new Error('Pastille did not report its hardware id. Update its firmware, then try again.');
	}
	return hardwareId;
}

async function configure(prov: RpcChannel, claimed: ClaimedDevice, options: PairOptions): Promise<void> {
	try {
		await prov.call('prov.configure', configureParams(claimed, options));
	} catch (error) {
		await revokeDevice(claimed.device.id).catch(() => undefined);
		throw error;
	}
}

function configureParams(claimed: ClaimedDevice, options: PairOptions): Record<string, unknown> {
	return {
		gateway_url: claimed.gateway_ws_url,
		device_token: claimed.device_token,
		...(options.wifiSsid ? { wifi_ssid: options.wifiSsid, wifi_psk: options.wifiPsk ?? '' } : {}),
		prefer_ble: true
	};
}

/**
 * Forwards frames between the device and the gateway. BLE notifications start
 * before the socket opens and every socket handler is attached at creation, so
 * the gateway's first tools/list is never missed. A socket that drops reopens
 * with backoff, except on 4401; a lost BLE link ends the relay.
 */
function startRelay(
	ble: BluetoothDevice,
	chars: Characteristics,
	write: FrameWriter,
	deviceId: string,
	options: RelayOptions
): { close(): void } {
	const report = (live: boolean, message: string) => options.onRelay?.(deviceId, { live, message });
	let socket: WebSocket | undefined;
	let stopped = false;
	let delay = RECONNECT_MIN_MS;
	let retry: ReturnType<typeof setTimeout> | undefined;
	let writeFailures = 0;

	const stop = (message: string) => {
		if (stopped) return;
		stopped = true;
		clearTimeout(retry);
		socket?.close(1000);
		ble.gatt?.disconnect();
		report(false, message);
	};
	const toDevice = (target: Characteristic, frame: Uint8Array, withResponse: boolean) =>
		write(target, frame, withResponse).then(
			() => (writeFailures = 0),
			() => {
				writeFailures += 1;
				if (writeFailures === WRITE_FAILURES_REPORTED) report(true, 'Bulle cannot reach your Pastille over Bluetooth.');
			}
		);
	const forward = (opcode: number) => (frame: Uint8Array) => {
		if (frame[0] === opcode && socket?.readyState === WebSocket.OPEN) socket.send(frame.slice());
	};
	const broken = () => stop('The Bluetooth stream from your Pastille broke. Connect it again.');

	const open = () => {
		const current = new WebSocket(relayUrl(deviceId), 'bohs.v1');
		socket = current;
		current.binaryType = 'arraybuffer';
		current.onopen = () => {
			delay = RECONNECT_MIN_MS;
			report(true, 'Relaying to Bulle');
		};
		current.onmessage = (event) => {
			const frame = new Uint8Array(event.data as ArrayBuffer);
			if (frame[0] === OPCODE_AUDIO) void toDevice(chars.audio, frame, false);
			if (frame[0] === OPCODE_CONTROL) void toDevice(chars.control, frame, true);
		};
		current.onerror = () => report(true, 'The relay to Bulle hit a network error.');
		current.onclose = (event) => {
			if (stopped || socket !== current) return;
			if (event.code === STATUS_REVOKED) return stop('This Pastille was revoked, or you signed out.');
			report(true, `The relay to Bulle closed. Retrying in ${Math.round(delay / 1000)} s.`);
			retry = setTimeout(open, delay);
			delay = Math.min(delay * 2, RECONNECT_MAX_MS);
		};
	};

	ble.addEventListener('gattserverdisconnected', () => stop('Your Pastille disconnected.'));
	listen(chars.audio, createReassembler(forward(OPCODE_AUDIO), broken));
	listen(chars.control, createReassembler(forward(OPCODE_CONTROL), broken));
	void Promise.all([chars.audio.startNotifications(), chars.control.startNotifications()]).then(
		() => {
			if (!stopped) open();
		},
		(error) => stop(readable(error).message)
	);
	return { close: () => stop('Relay stopped.') };
}

function readable(error: unknown): Error {
	if (isSecurityError(error)) {
		return new Error(
			'Pastille only accepts an encrypted Bluetooth link. Pair it in your system Bluetooth settings, then try again.'
		);
	}
	return error instanceof Error ? error : new Error(String(error));
}
