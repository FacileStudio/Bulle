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

function silent(signal: AbortSignal, frames: string[]): Response {
	const encoder = new TextEncoder();
	const body = new ReadableStream<Uint8Array>({
		start(controller) {
			for (const frame of frames) controller.enqueue(encoder.encode(frame));
			signal.addEventListener('abort', () => controller.error(new DOMException('aborted', 'AbortError')));
		}
	});
	return new Response(body, { status: 200 });
}

describe('LiveStream watchdog', () => {
	it('reconnects when a connection stays silent for 2.5 pings', async () => {
		let connects = 0;
		let opens = 0;
		let stream: LiveStream;
		const done = new Promise<void>((resolve) => {
			stream = new LiveStream({
				pingMs: 4,
				connect: async (signal) => {
					connects++;
					if (connects === 2) {
						stream.stop();
						resolve();
					}
					return silent(signal, [': ok\n\n']);
				},
				wait: async () => undefined
			});
		});
		stream!.listen({ onOpen: () => opens++ });
		const started = Date.now();
		stream!.start();
		await done;
		expect(connects).toBe(2);
		expect(opens).toBe(1);
		expect(Date.now() - started).toBeGreaterThanOrEqual(9);
	});

	it('stops and signs out on a 401', async () => {
		let connects = 0;
		const signedOut = new Promise<void>((resolve) => {
			const stream = new LiveStream({
				connect: async () => {
					connects++;
					return new Response(null, { status: 401 });
				},
				wait: async () => undefined
			});
			stream.listen({
				onUnauthorized: () => {
					expect(stream.running).toBe(false);
					resolve();
				}
			});
			stream.start();
		});
		await signedOut;
		expect(connects).toBe(1);
	});

	it('revives a stale stream and leaves a fresh one alone', async () => {
		let clock = 0;
		const signals: AbortSignal[] = [];
		const stream = new LiveStream({
			pingMs: 1000,
			now: () => clock,
			connect: async (signal) => {
				signals.push(signal);
				return silent(signal, [': ok\n\n']);
			},
			wait: () => new Promise(() => undefined)
		});
		stream.start();
		await new Promise((resolve) => setTimeout(resolve, 5));
		stream.revive();
		expect(signals).toHaveLength(1);
		clock = 5000;
		stream.revive();
		await new Promise((resolve) => setTimeout(resolve, 5));
		expect(signals).toHaveLength(2);
		expect(signals[0].aborted).toBe(true);
		stream.stop();
	});
});
