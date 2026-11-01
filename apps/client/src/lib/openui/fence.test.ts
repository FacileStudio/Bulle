import { describe, expect, it } from 'vitest';
import { splitFences } from './fence';

describe('splitFences', () => {
	it('returns plain markdown untouched', () => {
		expect(splitFences('hello\n\nworld')).toEqual([{ type: 'markdown', text: 'hello\n\nworld' }]);
	});

	it('splits markdown around a complete block', () => {
		const text = 'before\n```openui\nroot = Card([])\n```\nafter';
		expect(splitFences(text)).toEqual([
			{ type: 'markdown', text: 'before' },
			{ type: 'openui', source: 'root = Card([])', complete: true },
			{ type: 'markdown', text: 'after' }
		]);
	});

	it('marks an unterminated block as partial', () => {
		const text = 'intro\n```openui\nroot = Card([a])\na = Sta';
		expect(splitFences(text)).toEqual([
			{ type: 'markdown', text: 'intro' },
			{ type: 'openui', source: 'root = Card([a])\na = Sta', complete: false }
		]);
	});

	it('accepts the openui-lang info string', () => {
		const [segment] = splitFences('```openui-lang\nroot = Card([])\n```');
		expect(segment).toEqual({ type: 'openui', source: 'root = Card([])', complete: true });
	});

	it('handles several blocks', () => {
		const text = '```openui\na\n```\nmid\n```openui\nb\n```';
		expect(splitFences(text).map((s) => s.type)).toEqual(['openui', 'markdown', 'openui']);
	});

	it('leaves other fenced code alone, even if it mentions openui', () => {
		const text = '```md\n```openui\nx\n```\n```';
		const segments = splitFences(text);
		expect(segments.every((s) => s.type === 'markdown')).toBe(true);
	});

	it('does not treat an indented-by-text mention as a fence', () => {
		const text = 'use ```openui``` to start';
		expect(splitFences(text)).toEqual([{ type: 'markdown', text }]);
	});

	it('holds back a partial opening fence while streaming', () => {
		expect(splitFences('text\n```ope', { streaming: true })).toEqual([
			{ type: 'markdown', text: 'text' }
		]);
		expect(splitFences('text\n```ope')).toEqual([{ type: 'markdown', text: 'text\n```ope' }]);
	});

	it('opens a block as soon as the info string is complete', () => {
		expect(splitFences('```openui', { streaming: true })).toEqual([
			{ type: 'openui', source: '', complete: false }
		]);
	});

	it('closes with a longer fence', () => {
		const [segment] = splitFences('```openui\nx\n````');
		expect(segment).toMatchObject({ complete: true, source: 'x' });
	});

	it('returns nothing for empty input', () => {
		expect(splitFences('')).toEqual([]);
	});
});
