import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { APK_URL } from '$lib/downloads';
import Page from './+page.svelte';

describe('downloads page', () => {
	it('links the latest APK and announces the desktop app', () => {
		const html = render(Page).body;
		expect(html).toContain(`href="${APK_URL}"`);
		expect(html).toContain('Coming soon');
		expect(html).toContain('facile install bulle');
	});
});
