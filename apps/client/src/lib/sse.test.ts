import { describe, expect, it } from 'vitest';
import { createSseParser, type SseFrame } from './sse';

function run(chunks: string[]): SseFrame[] {
	const frames: SseFrame[] = [];
	const push = createSseParser((f) => frames.push(f));
	for (const chunk of chunks) push(chunk);
	return frames;
}

describe('createSseParser', () => {
	it('parses a whole frame', () => {
		expect(run(['id: 7\nevent: RUN_STARTED\ndata: {"a":1}\n\n'])).toEqual([
			{ id: '7', event: 'RUN_STARTED', data: '{"a":1}' }
		]);
	});

	it('handles chunks split anywhere', () => {
		const text = 'id: 1\nevent: THINKING\ndata: {"delta":"hi"}\n\nid: 2\nevent: RUN_ERROR\ndata: x\n\n';
		for (let i = 1; i < text.length; i++) {
			const frames = run([text.slice(0, i), text.slice(i)]);
			expect(frames.map((f) => f.id)).toEqual(['1', '2']);
			expect(frames[0].data).toBe('{"delta":"hi"}');
		}
	});

	it('handles CRLF split between chunks', () => {
		expect(run(['id: 3\r', '\nevent: A\r\ndata: b\r', '\n\r', '\n'])).toEqual([
			{ id: '3', event: 'A', data: 'b' }
		]);
	});

	it('joins multi-line data and skips comments', () => {
		expect(run([': ping\n\ndata: a\ndata: b\n\n'])).toEqual([{ id: '', event: '', data: 'a\nb' }]);
	});
});
