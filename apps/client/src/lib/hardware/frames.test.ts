import { describe, expect, it } from 'vitest';
import {
	MAX_FRAME,
	OPCODE_AUDIO,
	OPCODE_PROVISION,
	chunkFrame,
	createReassembler,
	jsonRpcFrame,
	parseJsonRpcFrame
} from './frames';

function audioFrame(seq: number): Uint8Array {
	const frame = new Uint8Array(4 + 640);
	frame[0] = OPCODE_AUDIO;
	new DataView(frame.buffer).setUint16(2, seq);
	frame.fill(seq & 0xff, 4);
	return frame;
}

describe('bohs.v1 BLE length-prefix codec', () => {
	it('prefixes the length and keeps every chunk within the MTU', () => {
		const chunks = chunkFrame(audioFrame(7), 244);
		expect(chunks.map((c) => c.length)).toEqual([244, 244, 158]);
		expect([chunks[0][0], chunks[0][1]]).toEqual([0x02, 0x84]);
	});

	it('reassembles frames split across and inside chunks', () => {
		const status = { jsonrpc: '2.0', id: 3, result: { state: 'authorized' } };
		const sent = [audioFrame(1), jsonRpcFrame(OPCODE_PROVISION, status), audioFrame(2)];
		const stream = sent.flatMap((frame) => chunkFrame(frame, 244));
		const bytes = new Uint8Array(stream.reduce((n, c) => n + c.length, 0));
		stream.reduce((at, c) => (bytes.set(c, at), at + c.length), 0);
		const received: Uint8Array[] = [];
		const feed = createReassembler((frame) => received.push(frame));
		for (let at = 0; at < bytes.length; at += 17) feed(bytes.subarray(at, at + 17));
		expect(received).toEqual(sent);
		expect(parseJsonRpcFrame(received[1])).toEqual(status);
	});

	it('refuses an empty frame or one above 1024 bytes', () => {
		expect(() => chunkFrame(new Uint8Array(MAX_FRAME + 1), 20)).toThrow(RangeError);
		expect(() => chunkFrame(new Uint8Array(0), 20)).toThrow(RangeError);
		expect(chunkFrame(new Uint8Array(MAX_FRAME), 244)).toHaveLength(5);
	});

	it('drops a stream whose length prefix is out of range, then reads the next frame', () => {
		const received: Uint8Array[] = [];
		let broken = 0;
		const feed = createReassembler(
			(frame) => received.push(frame),
			() => broken++
		);
		feed(new Uint8Array([0x04, 0x01, 0xaa, 0xbb]));
		feed(new Uint8Array([0x00, 0x00]));
		const next = audioFrame(3);
		chunkFrame(next, 244).forEach(feed);
		expect(broken).toBe(2);
		expect(received).toEqual([next]);
	});
});
