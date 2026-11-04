export const OPCODE_AUDIO = 0x01;
export const OPCODE_CONTROL = 0x02;
export const OPCODE_PROVISION = 0x03;

const MAX_FRAME = 0xffff;

/**
 * Splits one bohs.v1 frame into BLE writes: a u16be length prefix, then the
 * prefixed bytes in chunks of at most `chunkSize` (ATT_MTU - 3).
 */
export function chunkFrame(frame: Uint8Array, chunkSize: number): Uint8Array[] {
	if (frame.length > MAX_FRAME) throw new RangeError('A bohs.v1 frame is at most 65535 bytes');
	if (!Number.isInteger(chunkSize) || chunkSize < 1) {
		throw new RangeError('chunkSize must be a positive integer');
	}
	const framed = new Uint8Array(frame.length + 2);
	new DataView(framed.buffer).setUint16(0, frame.length);
	framed.set(frame, 2);
	const chunks: Uint8Array[] = [];
	for (let at = 0; at < framed.length; at += chunkSize) {
		chunks.push(framed.subarray(at, at + chunkSize));
	}
	return chunks;
}

/**
 * Answers a feeder for BLE notification chunks. It reassembles the
 * length-prefixed stream and calls `onFrame` once per whole frame, however the
 * chunks split it.
 */
export function createReassembler(
	onFrame: (frame: Uint8Array) => void
): (chunk: Uint8Array) => void {
	let buffer = new Uint8Array(0);
	return (chunk) => {
		const joined = new Uint8Array(buffer.length + chunk.length);
		joined.set(buffer);
		joined.set(chunk, buffer.length);
		buffer = joined;
		while (buffer.length >= 2) {
			const length = (buffer[0] << 8) | buffer[1];
			if (buffer.length < length + 2) return;
			onFrame(buffer.slice(2, 2 + length));
			buffer = buffer.slice(2 + length);
		}
	};
}

/** Encodes a JSON-RPC message as a control (0x02) or provisioning (0x03) frame. */
export function jsonRpcFrame(opcode: number, message: unknown): Uint8Array {
	const body = new TextEncoder().encode(JSON.stringify(message));
	const frame = new Uint8Array(body.length + 1);
	frame[0] = opcode;
	frame.set(body, 1);
	return frame;
}

/** Decodes the JSON-RPC message of a control or provisioning frame, or answers undefined. */
export function parseJsonRpcFrame(frame: Uint8Array): unknown {
	if (frame.length < 2) return undefined;
	try {
		return JSON.parse(new TextDecoder().decode(frame.subarray(1)));
	} catch {
		return undefined;
	}
}
