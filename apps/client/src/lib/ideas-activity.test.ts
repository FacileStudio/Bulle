import { afterEach, describe, expect, it, vi } from 'vitest';
import { ideasActivity, refreshIdeas } from './ideas-activity.svelte';
import { spaces } from './spaces.svelte';

function overview(title: string) {
	return {
		ideas: [{ id: title, title, status: 'new', created_at: '', expires_at: '' }],
		goals: [],
		settings: { space_id: title, enabled: true, hour: 8, timezone: 'UTC', push: false, steering: '' },
		generating: false,
		last_batch_at: null
	};
}

describe('refreshIdeas', () => {
	const originalFetch = globalThis.fetch;

	afterEach(() => {
		globalThis.fetch = originalFetch;
	});

	it('drops a slow answer for a space the user already left', async () => {
		let answerA: (value: Response) => void = () => undefined;
		globalThis.fetch = vi
			.fn()
			.mockImplementationOnce(() => new Promise<Response>((resolve) => (answerA = resolve)))
			.mockResolvedValueOnce({ ok: true, status: 200, json: async () => overview('b') } as Response);

		spaces.currentId = 'a';
		const first = refreshIdeas();
		spaces.currentId = 'b';
		await refreshIdeas();
		answerA({ ok: true, status: 200, json: async () => overview('a') } as Response);
		await first;

		expect(ideasActivity.spaceId).toBe('b');
		expect(ideasActivity.ideas.map((idea) => idea.id)).toEqual(['b']);
	});

	it('keeps the newest answer when an older poll for the same space lands last', async () => {
		let answerOld: (value: Response) => void = () => undefined;
		globalThis.fetch = vi
			.fn()
			.mockImplementationOnce(() => new Promise<Response>((resolve) => (answerOld = resolve)))
			.mockResolvedValueOnce({ ok: true, status: 200, json: async () => overview('new') } as Response);

		spaces.currentId = 'c';
		const older = refreshIdeas();
		await refreshIdeas();
		answerOld({ ok: true, status: 200, json: async () => overview('old') } as Response);
		await older;

		expect(ideasActivity.ideas.map((idea) => idea.id)).toEqual(['new']);
	});
});
