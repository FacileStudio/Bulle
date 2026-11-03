import { describe, expect, it } from 'vitest';
import { parseAllow, statusTone } from './matrix';

describe('matrix settings helpers', () => {
	it('splits the allowlist on commas, spaces and new lines', () => {
		expect(parseAllow(' @a:m.org, @b:m.org\n\n@c:m.org ')).toEqual(['@a:m.org', '@b:m.org', '@c:m.org']);
		expect(parseAllow('')).toEqual([]);
	});

	it('maps the bot status to a badge tone', () => {
		expect(statusTone('running')).toBe('success');
		expect(statusTone('connecting')).toBe('info');
		expect(statusTone('error: login failed')).toBe('danger');
		expect(statusTone('off')).toBe('neutral');
	});
});
