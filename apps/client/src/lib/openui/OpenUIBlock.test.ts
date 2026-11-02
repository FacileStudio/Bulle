import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import OpenUIBlock from './OpenUIBlock.svelte';

const html = (source: string, streaming = false) =>
	render(OpenUIBlock, { props: { source, streaming } }).body;

describe('OpenUIBlock', () => {
	it('renders whitelisted components', () => {
		const out = html(
			[
				'root = Card([s, t, txt, d, f], "Report", "Monthly overview")',
				's = Stat("Users", 12, "+3")',
				't = Table(["a"], [["cell-value"]])',
				'txt = Text("Important notes here", "muted")',
				'd = Divider()',
				'f = Form("signup", "Send", [e])',
				'e = Field("email", "Email")'
			].join('\n')
		);
		expect(out).toContain('Report');
		expect(out).toContain('Monthly overview');
		expect(out).toContain('Users');
		expect(out).toContain('cell-value');
		expect(out).toContain('Important notes here');
		expect(out).toContain('Send');
		expect(out).not.toContain('data-openui-fallback');
	});

	it('shows the raw source for an unknown component', () => {
		const source = 'root = Card([x])\nx = Iframe("https://evil.example")';
		const out = html(source);
		expect(out).toContain('data-openui-fallback');
		expect(out).toContain('Iframe');
		expect(out).not.toContain('<iframe');
	});

	it('shows the raw source for unparseable text', () => {
		const out = html('<script>alert(1)</script>');
		expect(out).toContain('data-openui-fallback');
		expect(out).not.toContain('<script>alert');
	});

	it('shows a placeholder while the first statement streams in', () => {
		const out = html('roo', true);
		expect(out).not.toContain('data-openui-fallback');
	});
});
