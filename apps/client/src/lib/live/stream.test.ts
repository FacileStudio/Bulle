import { describe, expect, it } from 'vitest';
import type { LiveEvent } from './events';
import { LiveStream } from './stream';

function sse(frames: string[]): Response {
	const encoder = new TextEncoder();
	const body = new ReadableStream<Uint8Array>({
		start(controller) {
			for (const frame of frames) controller.enqueue(encoder.encode(frame));
			controller.close();
		}
	});
	return new Response(body, { status: 200 });
}

describe('LiveStream', () => {
	it('delivers events and resyncs on every reconnect with backoff', async () => {
		const responses = [
			sse([': ok\n\n', 'event: session.deleted\ndata: {"session_id":"a"}\n\n']),
			new Response(null, { status: 502 }),
			sse([': ok\n\n', 'event: run.started\ndata: {"session_id":"b","run_id":"r","session_status":"running"}\n\n'])
		];
		const waits: number[] = [];
		const events: LiveEvent[] = [];
		let opens = 0;
		let stream: LiveStream;
		const done = new Promise<void>((resolve) => {
			stream = new LiveStream({
				connect: async () => responses.shift() ?? new Response(null, { status: 503 }),
				wait: async (ms) => {
					waits.push(ms);
					if (waits.length === 3) {
						stream.stop();
						resolve();
					}
				}
			});
		});
		stream!.listen({ onOpen: () => opens++, onEvent: (e) => events.push(e) });
		stream!.start();
		stream!.start();
		await done;
		expect(opens).toBe(2);
		expect(events.map((e) => e.type)).toEqual(['session.deleted', 'run.started']);
		expect(waits).toEqual([1000, 2000, 1000]);
		expect(stream!.running).toBe(false);
	});
});
