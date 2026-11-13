import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { DOWNLOADS, INSTALL_METHODS } from '$lib/downloads';
import Page from './+page.svelte';

describe('downloads page', () => {
	it('links every release asset and the CLI install commands', () => {
		const html = render(Page).body;
		for (const entry of DOWNLOADS) expect(html).toContain(`href="${entry.href}"`);
		for (const method of INSTALL_METHODS) expect(html).toContain(method.command.replaceAll('&', '&amp;'));
	});
});
