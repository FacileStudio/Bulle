import { describe, expect, it } from 'vitest';
import { base64Url, parseDeepLink, pkceChallenge, randomString } from './platform';

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
