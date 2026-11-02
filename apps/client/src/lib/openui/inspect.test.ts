import { describe, expect, it } from 'vitest';
import { createInspector } from './inspect';
import { COMPONENT_NAMES, library } from './library';

const inspect = createInspector(library);

const valid = [
	'root = Card([s, t], "Title")',
	's = Stat("Users", 12, "+3")',
	't = Table(["a", "b"], [[1, 2], [3, 4]])'
].join('\n');

describe('library', () => {
	it('exposes exactly the whitelisted components', () => {
		expect([...COMPONENT_NAMES].sort()).toEqual(
			[
				'Badge',
				'Button',
				'Callout',
				'Card',
				'Chart',
				'Divider',
				'Field',
				'Form',
				'List',
				'Stat',
				'Table',
				'Text'
			].sort()
		);
	});
});

describe('inspect', () => {
	it('accepts a valid program', () => {
		expect(inspect(valid, false)).toEqual({ status: 'ready' });
	});

	it('accepts a truncated program while streaming', () => {
		const partial = valid.slice(0, valid.length - 12);
		expect(inspect(partial, true).status).not.toBe('fallback');
	});

	it('is pending before any statement has arrived', () => {
		expect(inspect('', true)).toEqual({ status: 'pending' });
		expect(inspect('roo', true).status).toBe('pending');
	});

	it('falls back on an unknown component', () => {
		const source = 'root = Card([x])\nx = Iframe("https://evil.example")';
		expect(inspect(source, false).status).toBe('fallback');
		expect(inspect(source, true).status).toBe('fallback');
	});

	it('falls back on garbage once the stream is done', () => {
		expect(inspect('this is not openui', false).status).toBe('fallback');
	});

	it('falls back on a missing required prop once complete', () => {
		expect(inspect('root = Card([s])\ns = Stat()', false).status).toBe('fallback');
	});

	it('never throws on hostile input', () => {
		for (const source of ['(((', 'root = ', '"unterminated', 'root = Card([[[', '\u0000']) {
			expect(() => inspect(source, false)).not.toThrow();
			expect(() => inspect(source, true)).not.toThrow();
		}
	});
});
