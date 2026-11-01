import { describe, expect, it } from 'vitest';
import { renderMarkdown } from './markdown';

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
