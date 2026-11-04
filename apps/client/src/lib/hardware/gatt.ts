import { OPCODE_PROVISION, chunkFrame, createReassembler, jsonRpcFrame, parseJsonRpcFrame } from './frames';

const PREFERRED_CHUNK = 244;
const SAFE_CHUNK = 20;
const RPC_TIMEOUT_MS = 5000;

export type Characteristic = {
	value?: DataView;
	readValue(): Promise<DataView>;
	writeValueWithResponse(value: Uint8Array): Promise<void>;
	writeValueWithoutResponse(value: Uint8Array): Promise<void>;
	startNotifications(): Promise<Characteristic>;
	addEventListener(type: 'characteristicvaluechanged', listener: (event: Event) => void): void;
};

export type FrameWriter = (
	characteristic: Characteristic,
	frame: Uint8Array,
	withResponse: boolean
) => Promise<void>;

type RpcMessage = {
	id?: number;
	result?: { state?: string };
	params?: { state?: string };
	error?: { code?: number; message?: string };
};

type Waiter = { resolve: (result: unknown) => void; reject: (error: Error) => void };

export type RpcChannel = {
	call(method: string, params?: Record<string, unknown>): Promise<unknown>;
	onState(listener: (state: string, message: RpcMessage) => void): () => void;
};

/**
 * Answers a writer that length-prefixes and chunks each frame, and runs GATT
 * writes one at a time. Chunks start at 244 bytes (ATT_MTU 247) and drop to 20
 * for good when the first chunk of a frame is refused as too long.
 */
export function createFrameWriter(): FrameWriter {
	let chunkSize = PREFERRED_CHUNK;
	let queue: Promise<void> = Promise.resolve();

	async function writeFrame(target: Characteristic, frame: Uint8Array, withResponse: boolean) {
		const chunks = chunkFrame(frame, chunkSize);
		for (const [index, chunk] of chunks.entries()) {
			try {
				await (withResponse ? target.writeValueWithResponse(chunk) : target.writeValueWithoutResponse(chunk));
			} catch (error) {
				if (index > 0 || chunkSize === SAFE_CHUNK || isSecurityError(error)) throw error;
				chunkSize = SAFE_CHUNK;
				return writeFrame(target, frame, withResponse);
			}
		}
	}

	return (target, frame, withResponse) => {
		const next = queue.then(() => writeFrame(target, frame, withResponse));
		queue = next.catch(() => undefined);
		return next;
	};
}

/** Feeds every notification or indication of the characteristic to `feed`. */
export function listen(target: Characteristic, feed: (chunk: Uint8Array) => void): void {
	target.addEventListener('characteristicvaluechanged', (event) => {
		const value = (event.target as unknown as Characteristic).value;
		if (value) feed(new Uint8Array(value.buffer, value.byteOffset, value.byteLength));
	});
}

/**
 * Opens a JSON-RPC 2.0 channel over the provisioning characteristic: 0x03
 * frames written with response, answers read from its indications.
 */
export async function openRpcChannel(target: Characteristic, write: FrameWriter): Promise<RpcChannel> {
	let nextId = 1;
	const pending = new Map<number, Waiter>();
	const stateListeners = new Set<(state: string, message: RpcMessage) => void>();

	listen(
		target,
		createReassembler((frame) => {
			if (frame[0] !== OPCODE_PROVISION) return;
			const message = parseJsonRpcFrame(frame) as RpcMessage | undefined;
			if (!message) return;
			const state = message.result?.state ?? message.params?.state;
			if (typeof state === 'string') stateListeners.forEach((listener) => listener(state, message));
			settle(pending, message);
		})
	);
	await target.startNotifications();

	return {
		async call(method, params = {}) {
			const id = nextId++;
			const reply = new Promise<unknown>((resolve, reject) => {
				pending.set(id, { resolve, reject });
				setTimeout(() => {
					if (pending.delete(id)) reject(new Error(`Pastille did not answer ${method}.`));
				}, RPC_TIMEOUT_MS);
			});
			await write(target, jsonRpcFrame(OPCODE_PROVISION, { jsonrpc: '2.0', id, method, params }), true);
			return reply;
		},
		onState(listener) {
			stateListeners.add(listener);
			return () => stateListeners.delete(listener);
		}
	};
}

/** Reports whether a Web Bluetooth error comes from the device refusing an unencrypted link. */
export function isSecurityError(error: unknown): boolean {
	if (!(error instanceof Error)) return false;
	if (error.name === 'SecurityError') return true;
	return error.name !== 'Error' && /secur|encrypt|authenticat|insufficient/i.test(error.message);
}

function settle(pending: Map<number, Waiter>, message: RpcMessage) {
	if (typeof message.id !== 'number') return;
	const waiter = pending.get(message.id);
	if (!waiter) return;
	pending.delete(message.id);
	if (message.error) {
		waiter.reject(new Error(message.error.message || `Pastille refused the request (${message.error.code}).`));
		return;
	}
	waiter.resolve(message.result);
}
