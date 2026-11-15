import { beforeEach, describe, expect, it, vi } from 'vitest';

const apiFetch = vi.hoisted(() => vi.fn());

vi.mock('$lib/backend', async (importOriginal) => ({
	...(await importOriginal<typeof import('./backend')>()),
	apiFetch: (...args: unknown[]) => apiFetch(...args)
}));

import { ApiError } from './backend';
import { createShare, getPublicShare, getShare, revokeShare, sharePath } from './shares';

beforeEach(() => apiFetch.mockReset());

describe('sharePath', () => {
	it('puts /share before the query that says where the artifact lives', () => {
		expect(sharePath('a b', { myceliumSpace: 'common' })).toBe('/artifacts/a%20b/share?mycelium_space=common');
		expect(sharePath('a', { myceliumSpace: 'sp', spaceId: 's1' })).toBe(
			'/artifacts/a/share?space_id=s1&mycelium_space=sp'
		);
		expect(sharePath('a', {})).toBe('/artifacts/a/share');
	});
});

describe('public links', () => {
	it('answers null when the artifact has no live link, and rethrows other failures', async () => {
		apiFetch.mockRejectedValueOnce(new ApiError('nope', 404));
		expect(await getShare('a', { myceliumSpace: 'common' })).toBeNull();
		apiFetch.mockRejectedValueOnce(new ApiError('down', 503));
		await expect(getShare('a', {})).rejects.toThrow('down');
	});

	it('creates with POST, revokes with DELETE and reads the public page by token', async () => {
		apiFetch.mockResolvedValue(undefined);
		await createShare('a', { myceliumSpace: 'common' });
		await revokeShare('a', { myceliumSpace: 'common' });
		await getPublicShare('t/k');
		expect(apiFetch.mock.calls).toEqual([
			['/artifacts/a/share?mycelium_space=common', { method: 'POST' }],
			['/artifacts/a/share?mycelium_space=common', { method: 'DELETE' }],
			['/public/shares/t%2Fk']
		]);
	});
});
