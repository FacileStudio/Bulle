import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { APK_URL, DESKTOP_BUILDS, detectDesktopOs, featuredPlatform } from '$lib/downloads';
import Page from './+page.svelte';

describe('downloads page', () => {
	it('links the latest APK, every desktop build and the CLI', () => {
		const html = render(Page).body;
		expect(html).toContain(`href="${APK_URL}"`);
		for (const build of DESKTOP_BUILDS) expect(html).toContain(`href="${build.url}"`);
		expect(html).not.toContain('Coming soon');
		expect(html).toContain('facile install bulle');
	});
});

describe('detectDesktopOs', () => {
	it('maps desktop agents and ignores phones', () => {
		expect(detectDesktopOs('Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5)')).toBe('macos');
		expect(detectDesktopOs('Mozilla/5.0 (Windows NT 10.0; Win64; x64)')).toBe('windows');
		expect(detectDesktopOs('Mozilla/5.0 (X11; Linux x86_64)')).toBe('linux');
		expect(detectDesktopOs('Mozilla/5.0 (Linux; Android 14; Pixel 8)')).toBeNull();
		expect(detectDesktopOs('Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)')).toBeNull();
	});
});

describe('featuredPlatform', () => {
	it('features the desktop OS, Android on a phone, and macOS otherwise', () => {
		expect(featuredPlatform('Mozilla/5.0 (Windows NT 10.0; Win64; x64)')).toBe('windows');
		expect(featuredPlatform('Mozilla/5.0 (Linux; Android 14; Pixel 8)')).toBe('android');
		expect(featuredPlatform('Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)')).toBe('macos');
	});
});
