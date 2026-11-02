import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { deleteCredential, listCredentials, putCredential, testCredential } from './api';

describe('credentials api client', () => {
	const originalFetch = globalThis.fetch;

	beforeEach(() => {
		globalThis.fetch = vi.fn();
	});

	afterEach(() => {
		globalThis.fetch = originalFetch;
	});

	it('listCredentials calls GET /api/credentials', async () => {
		const rows = [{ backend: 'openai', last4: 'abcd', updated_at: '', source: 'user' }];
		(globalThis.fetch as any).mockResolvedValueOnce({ ok: true, status: 200, json: async () => rows });

		expect(await listCredentials()).toEqual(rows);
		expect(globalThis.fetch).toHaveBeenCalledWith(
			'/api/credentials',
			expect.objectContaining({ credentials: 'same-origin' })
		);
	});

	it('putCredential sends PUT with the api_key body', async () => {
		(globalThis.fetch as any).mockResolvedValueOnce({ ok: true, status: 204 });

		await putCredential('anthropic', 'sk-test');
		const [url, init] = (globalThis.fetch as any).mock.calls[0];
		expect(url).toBe('/api/credentials/anthropic');
		expect(init.method).toBe('PUT');
		expect(JSON.parse(init.body)).toEqual({ api_key: 'sk-test' });
	});

	it('deleteCredential sends DELETE', async () => {
		(globalThis.fetch as any).mockResolvedValueOnce({ ok: true, status: 204 });

		await deleteCredential('google');
		const [url, init] = (globalThis.fetch as any).mock.calls[0];
		expect(url).toBe('/api/credentials/google');
		expect(init.method).toBe('DELETE');
	});

	it('testCredential sends POST and returns the result', async () => {
		(globalThis.fetch as any).mockResolvedValueOnce({
			ok: true,
			status: 200,
			json: async () => ({ ok: false, error: 'bad key' })
		});

		expect(await testCredential('openrouter')).toEqual({ ok: false, error: 'bad key' });
		const [url, init] = (globalThis.fetch as any).mock.calls[0];
		expect(url).toBe('/api/credentials/openrouter/test');
		expect(init.method).toBe('POST');
	});
});
