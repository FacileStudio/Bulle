import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { badgeFor, getSuiteStatus } from './suite';

describe('suite status client', () => {
	const originalFetch = globalThis.fetch;

	beforeEach(() => {
		globalThis.fetch = vi.fn();
	});

	afterEach(() => {
		globalThis.fetch = originalFetch;
	});

	it('reads the apps from GET /api/suite/status', async () => {
		const apps = [{ app: 'courrier', name: 'Courrier', url: 'https://c.test', state: 'connected' }];
		(globalThis.fetch as any).mockResolvedValueOnce({ ok: true, status: 200, json: async () => ({ apps }) });
		expect(await getSuiteStatus()).toEqual(apps);
		expect(globalThis.fetch).toHaveBeenCalledWith(
			'/api/suite/status',
			expect.objectContaining({ credentials: 'same-origin' })
		);
	});

	it('treats a missing route as no configured apps', async () => {
		(globalThis.fetch as any).mockResolvedValueOnce({ ok: false, status: 404, json: async () => ({}) });
		expect(await getSuiteStatus()).toEqual([]);
	});

	it('lets other failures through', async () => {
		(globalThis.fetch as any).mockResolvedValueOnce({ ok: false, status: 500, json: async () => ({}) });
		await expect(getSuiteStatus()).rejects.toThrow();
	});

	it('maps each state to a badge and falls back for unknown ones', () => {
		expect(badgeFor('connected').tone).toBe('success');
		expect(badgeFor('reconnect').label).toBe('Reconnect');
		expect(badgeFor('weird' as any).tone).toBe('neutral');
	});
});
