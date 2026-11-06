import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from './backend';
import {
	countNew,
	errorText,
	generateIdeas,
	getIdeas,
	groupIdeas,
	isActive,
	sortGoals,
	timezoneToSync,
	updateIdea,
	type Goal,
	type Idea
} from './ideas';

const NOW = Date.parse('2026-11-06T09:00:00Z');

function idea(id: string, status: Idea['status'], created: string, expires = '2026-11-08T09:00:00Z'): Idea {
	return { id, status, created_at: created, expires_at: expires } as Idea;
}

describe('isActive', () => {
	it('keeps new ideas until they expire and saved ones regardless', () => {
		expect(isActive(idea('a', 'new', '2026-11-06T08:00:00Z'), NOW)).toBe(true);
		expect(isActive(idea('b', 'new', '2026-11-01T08:00:00Z', '2026-11-04T08:00:00Z'), NOW)).toBe(false);
		expect(isActive(idea('c', 'saved', '2026-11-01T08:00:00Z', '2026-11-04T08:00:00Z'), NOW)).toBe(true);
		expect(isActive(idea('d', 'dismissed', '2026-11-06T08:00:00Z'), NOW)).toBe(false);
		expect(isActive(idea('e', 'done', '2026-11-06T08:00:00Z'), NOW)).toBe(false);
	});
});

describe('groupIdeas and countNew', () => {
	const ideas = [
		idea('old', 'new', '2026-11-05T08:00:00Z'),
		idea('fresh', 'new', '2026-11-06T08:00:00Z'),
		idea('kept', 'saved', '2026-11-04T08:00:00Z'),
		idea('stale', 'new', '2026-11-02T08:00:00Z', '2026-11-05T08:00:00Z'),
		idea('gone', 'dismissed', '2026-11-06T08:00:00Z')
	];

	it('puts the newest batch first and saved ideas apart', () => {
		const groups = groupIdeas(ideas, NOW);
		expect(groups.fresh.map((i) => i.id)).toEqual(['fresh', 'old']);
		expect(groups.saved.map((i) => i.id)).toEqual(['kept']);
	});

	it('counts only active new ideas for the dot', () => {
		expect(countNew(ideas, NOW)).toBe(2);
		expect(countNew([], NOW)).toBe(0);
	});
});

describe('sortGoals', () => {
	it('lists open goals before done ones', () => {
		const goals = [
			{ id: 'x', done: true, created_at: '2026-11-01' },
			{ id: 'y', done: false, created_at: '2026-11-03' },
			{ id: 'z', done: false, created_at: '2026-11-02' }
		] as Goal[];
		expect(sortGoals(goals).map((g) => g.id)).toEqual(['z', 'y', 'x']);
	});
});

describe('timezoneToSync', () => {
	it('sends the browser zone only over the UTC default', () => {
		expect(timezoneToSync('UTC', 'Europe/Paris')).toBe('Europe/Paris');
		expect(timezoneToSync('UTC', 'UTC')).toBeNull();
		expect(timezoneToSync('America/New_York', 'Europe/Paris')).toBeNull();
		expect(timezoneToSync('UTC', undefined)).toBeNull();
	});
});

describe('ideas api client', () => {
	const originalFetch = globalThis.fetch;
	const ok = (body: unknown) => ({ ok: true, status: 200, json: async () => body });

	beforeEach(() => {
		globalThis.fetch = vi.fn();
	});

	afterEach(() => {
		globalThis.fetch = originalFetch;
	});

	it('scopes the list to the space', async () => {
		vi.mocked(globalThis.fetch).mockResolvedValueOnce(ok({ ideas: [] }) as Response);
		await getIdeas('sp 1');
		expect(globalThis.fetch).toHaveBeenCalledWith('/api/ideas?space_id=sp%201', expect.anything());
	});

	it('patches an idea with the CSRF header', async () => {
		vi.mocked(globalThis.fetch).mockResolvedValueOnce(ok({ id: 'i1' }) as Response);
		await updateIdea('i1', { status: 'saved' });
		const [url, init] = vi.mocked(globalThis.fetch).mock.calls[0];
		expect(url).toBe('/api/ideas/i1');
		expect(init?.method).toBe('PATCH');
		expect(init?.body).toBe(JSON.stringify({ status: 'saved' }));
		expect(new Headers(init?.headers).get('X-Facile-CSRF')).toBe('1');
	});

	it('surfaces a 429 status on generate', async () => {
		vi.mocked(globalThis.fetch).mockResolvedValueOnce({
			ok: false,
			status: 429,
			json: async () => ({ error: { message: 'at most 3 manual batches per day' } })
		} as Response);
		await expect(generateIdeas('s1')).rejects.toMatchObject({
			status: 429,
			message: 'at most 3 manual batches per day'
		});
	});
});

describe('errorText', () => {
	it('shows the server reason for a 4xx and the fallback otherwise', () => {
		expect(errorText(new ApiError('goal title must be 1 to 120 characters', 400), 'x')).toBe(
			'Goal title must be 1 to 120 characters.'
		);
		expect(errorText(new ApiError('failed to load idea', 500), 'Could not.')).toBe('Could not.');
		expect(errorText(new TypeError('Failed to fetch'), 'Could not.')).toBe('Could not.');
	});
});
