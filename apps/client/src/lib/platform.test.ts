import { beforeEach, describe, expect, it, vi } from 'vitest';
import { base64Url, parseDeepLink, pkceChallenge, randomState, randomString } from './platform';

const fake = vi.hoisted(() => ({
	native: false,
	desktop: false,
	refuse: false,
	prefs: new Map<string, string>(),
	secure: new Map<string, string>(),
	keychain: new Map<string, string>()
}));

vi.mock('@capacitor/core', () => ({
	Capacitor: { isNativePlatform: () => fake.native }
}));

vi.mock('@tauri-apps/api/core', () => ({
	isTauri: () => fake.desktop,
	invoke: async (command: string, args?: { token: string }) => {
		if (fake.refuse) throw new Error('no keychain');
		if (command === 'token_get') return fake.keychain.get('session') ?? null;
		if (command === 'token_set' && args) fake.keychain.set('session', args.token);
		if (command === 'token_clear') fake.keychain.delete('session');
		return null;
	}
}));

vi.mock('capacitor-secure-storage-plugin', () => ({
	SecureStoragePlugin: {
		get: async ({ key }: { key: string }) => {
			const value = fake.secure.get(key);
			if (value === undefined) throw new Error('Item with given key does not exist');
			return { value };
		},
		set: async ({ key, value }: { key: string; value: string }) => {
			if (fake.refuse) throw new Error('keystore unavailable');
			fake.secure.set(key, value);
			return { value: true };
		},
		remove: async ({ key }: { key: string }) => {
			if (!fake.secure.delete(key)) throw new Error('Item with given key does not exist');
			return { value: true };
		}
	}
}));

vi.mock('@capacitor/preferences', () => ({
	Preferences: {
		get: async ({ key }: { key: string }) => ({ value: fake.prefs.get(key) ?? null }),
		set: async ({ key, value }: { key: string; value: string }) => void fake.prefs.set(key, value),
		remove: async ({ key }: { key: string }) => void fake.prefs.delete(key)
	}
}));

async function freshPlatform() {
	vi.resetModules();
	return import('./platform');
}

describe('token storage', () => {
	beforeEach(() => {
		Object.assign(fake, { native: false, desktop: false, refuse: false });
		fake.prefs.clear();
		fake.secure.clear();
		fake.keychain.clear();
		vi.spyOn(console, 'warn').mockImplementation(() => undefined);
	});

	it('moves an Android token from Preferences into the secure store', async () => {
		fake.native = true;
		fake.prefs.set('bulle.token', 'old-token');
		const platform = await freshPlatform();
		expect(await platform.getToken()).toBe('old-token');
		expect(fake.secure.get('bulle.token')).toBe('old-token');
		expect(fake.prefs.has('bulle.token')).toBe(false);
		expect(await (await freshPlatform()).getToken()).toBe('old-token');
	});

	it('moves a desktop token from Preferences into the keychain', async () => {
		fake.desktop = true;
		fake.prefs.set('bulle.token', 'old-token');
		const platform = await freshPlatform();
		expect(await platform.getToken()).toBe('old-token');
		expect(fake.keychain.get('session')).toBe('old-token');
		expect(fake.prefs.has('bulle.token')).toBe(false);
	});

	it('keeps the Preferences copy when the secure store refuses it', async () => {
		fake.native = true;
		fake.refuse = true;
		fake.prefs.set('bulle.token', 'old-token');
		const platform = await freshPlatform();
		expect(await platform.getToken()).toBe('old-token');
		expect(fake.prefs.get('bulle.token')).toBe('old-token');
	});

	it('prefers the secure store and leaves Preferences alone', async () => {
		fake.native = true;
		fake.secure.set('bulle.token', 'new-token');
		fake.prefs.set('bulle.token', 'stale-token');
		const platform = await freshPlatform();
		expect(await platform.getToken()).toBe('new-token');
		expect(fake.prefs.get('bulle.token')).toBe('stale-token');
	});

	it('writes a new sign-in to the secure store only', async () => {
		fake.native = true;
		const platform = await freshPlatform();
		await platform.setToken('fresh-token');
		expect(fake.secure.get('bulle.token')).toBe('fresh-token');
		expect(fake.prefs.size).toBe(0);
	});

	it('signs out of both stores so a refused migration cannot sign back in', async () => {
		fake.native = true;
		fake.refuse = true;
		fake.prefs.set('bulle.token', 'old-token');
		const platform = await freshPlatform();
		await platform.getToken();
		await platform.clearToken();
		expect(await platform.getToken()).toBeNull();
		expect(fake.prefs.has('bulle.token')).toBe(false);
		expect(await (await freshPlatform()).getToken()).toBeNull();
	});

	it('keeps no token on the web', async () => {
		fake.prefs.set('bulle.token', 'old-token');
		const platform = await freshPlatform();
		expect(await platform.getToken()).toBeNull();
		await platform.setToken('web-token');
		expect(fake.secure.size + fake.keychain.size).toBe(0);
		expect(fake.prefs.get('bulle.token')).toBe('old-token');
	});
});

describe('pkce', () => {
	it('matches the RFC 7636 appendix B vector', async () => {
		const challenge = await pkceChallenge('dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk');
		expect(challenge).toBe('E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM');
	});

	it('encodes base64url without padding', () => {
		expect(base64Url(new Uint8Array([251, 255, 254]))).toBe('-__-');
		expect(base64Url(new Uint8Array([1]))).toBe('AQ');
	});

	it('builds verifiers of unreserved characters', () => {
		const verifier = randomString(64);
		expect(verifier).toMatch(/^[A-Za-z0-9\-._~]{64}$/);
	});
});

describe('randomState', () => {
	it('only uses characters porte keeps in the mobile state', () => {
		for (let i = 0; i < 200; i++) expect(randomState()).toMatch(/^[A-Za-z0-9_-]{32}$/);
	});
});

describe('parseDeepLink', () => {
	it('parses the auth callback', () => {
		expect(parseDeepLink('studio.facile.bulle://auth?code=abc&state=xyz')).toEqual({
			kind: 'auth',
			code: 'abc',
			state: 'xyz',
			error: ''
		});
	});

	it('parses a session link', () => {
		expect(parseDeepLink('studio.facile.bulle://sessions/s_42')).toEqual({
			kind: 'session',
			id: 's_42'
		});
	});

	it('rejects foreign schemes and unknown hosts', () => {
		expect(parseDeepLink('https://auth?code=1')).toBeNull();
		expect(parseDeepLink('studio.facile.bulle://other')).toBeNull();
		expect(parseDeepLink('studio.facile.bulle://sessions/')).toBeNull();
		expect(parseDeepLink('nonsense')).toBeNull();
	});
});
