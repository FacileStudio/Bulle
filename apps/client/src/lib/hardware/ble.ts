import { claimDevice, relayUrl, type ClaimedDevice, type HardwareDevice } from './api';
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

type GattServer = {
	getPrimaryService(uuid: string): Promise<{ getCharacteristic(uuid: string): Promise<Characteristic> }>;
};

type BluetoothDevice = {
	id: string;
	name?: string;
	gatt?: { connect(): Promise<GattServer>; disconnect(): void };
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

export type PairOptions = {
	name?: string;
	wifiSsid?: string;
	wifiPsk?: string;
	onStatus?: (message: string) => void;
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
	if (!bluetoothAvailable()) throw new Error('This browser has no Web Bluetooth. Use Chrome or Edge.');
	const bluetooth = (navigator as unknown as { bluetooth: Bluetooth }).bluetooth;
	const ble = await bluetooth.requestDevice({ filters: [{ services: [BOHS_SERVICE] }] }).catch(() => {
		throw new Error('No Pastille was selected.');
	});
	try {
		status('Connecting to your Pastille');
		const chars = await connect(ble);
		const info = parseInfo(await chars.info.readValue());
		const write = createFrameWriter();
		const prov = await openRpcChannel(chars.prov, write);
		await prov.call('prov.identify');
		status('Tap your Pastille to confirm pairing');
		const authorized = await waitForAuthorization(prov);
		status('Adding it to your account');
		const claimed = await claimDevice(hardwareIdOf(authorized, ble), options.name || ble.name);
		await prov.call('prov.configure', configureParams(claimed, options));
		const relay = await startRelay(chars, write, claimed.device.id, status);
		status('Paired');
		return {
			device: claimed.device,
			info,
			close() {
				relay.close();
				ble.gatt?.disconnect();
			}
		};
	} catch (error) {
		ble.gatt?.disconnect();
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

function hardwareIdOf(status: unknown, ble: BluetoothDevice): string {
	const reported = (status as { hardware_id?: unknown } | undefined)?.hardware_id;
	if (typeof reported === 'string' && reported) return reported;
	return Array.from(atob(ble.id), (c) => c.charCodeAt(0).toString(16).padStart(2, '0'))
		.join('')
		.toUpperCase();
}

function configureParams(claimed: ClaimedDevice, options: PairOptions): Record<string, unknown> {
	return {
		gateway_url: claimed.gateway_ws_url,
		device_token: claimed.device_token,
		...(options.wifiSsid ? { wifi_ssid: options.wifiSsid, wifi_psk: options.wifiPsk ?? '' } : {}),
		prefer_ble: true
	};
}

async function startRelay(
	chars: Characteristics,
	write: FrameWriter,
	deviceId: string,
	status: (message: string) => void
): Promise<{ close(): void }> {
	const socket = new WebSocket(relayUrl(deviceId), 'bohs.v1');
	socket.binaryType = 'arraybuffer';
	const forward = (frame: Uint8Array) => {
		if (socket.readyState === WebSocket.OPEN) socket.send(frame.slice());
	};
	listen(chars.audio, createReassembler(forward));
	listen(chars.control, createReassembler(forward));
	await Promise.all([chars.audio.startNotifications(), chars.control.startNotifications()]);
	socket.onmessage = (event) => {
		const frame = new Uint8Array(event.data as ArrayBuffer);
		if (frame[0] === OPCODE_AUDIO) void write(chars.audio, frame, false).catch(() => undefined);
		if (frame[0] === OPCODE_CONTROL) void write(chars.control, frame, true).catch(() => undefined);
	};
	socket.onclose = (event) =>
		status(event.code === STATUS_REVOKED ? 'This Pastille was revoked.' : 'The relay to Bulle closed.');
	return { close: () => socket.close(1000) };
}

function readable(error: unknown): Error {
	if (isSecurityError(error)) {
		return new Error(
			'Pastille only accepts an encrypted Bluetooth link. Pair it in your system Bluetooth settings, then try again.'
		);
	}
	return error instanceof Error ? error : new Error(String(error));
}
