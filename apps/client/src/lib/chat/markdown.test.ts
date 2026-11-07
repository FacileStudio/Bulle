import { describe, expect, it } from 'vitest';
import { renderMarkdown, splitSettled } from './markdown';

describe('splitSettled', () => {
	it('cuts at the last blank line outside a fence', () => {
		expect(splitSettled('one\n\ntwo\n\nthr')).toEqual({ head: 'one\n\ntwo\n\n', tail: 'thr' });
		expect(splitSettled('a\n\n```\nx\n\ny')).toEqual({ head: 'a\n\n', tail: '```\nx\n\ny' });
		expect(splitSettled('```\nx\n```\n\nz')).toEqual({ head: '```\nx\n```\n\n', tail: 'z' });
		expect(splitSettled('plain')).toEqual({ head: '', tail: 'plain' });
	});
});

describe('renderMarkdown', () => {
	it('renders basic markdown', () => {
		expect(renderMarkdown('**hi**')).toContain('<strong>hi</strong>');
	});

	it('escapes raw html', () => {
		const out = renderMarkdown('<script>alert(1)</script>');
		expect(out).not.toContain('<script>');
	});

	it('drops unsafe link schemes and hardens safe ones', () => {
		expect(renderMarkdown('[x](javascript:alert(1))')).not.toContain('href');
		const ok = renderMarkdown('[x](https://example.com)');
		expect(ok).toContain('rel="noopener noreferrer nofollow"');
	});
});
