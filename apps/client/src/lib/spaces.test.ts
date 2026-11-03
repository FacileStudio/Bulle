import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { listJobs } from './api';
import { draftFor, inputFromDraft } from './space-draft';
import { pickSpace, withSpace, type Space } from './spaces.svelte';

function space(id: string, extra: Partial<Space> = {}): Space {
	return {
		id,
		name: id,
		icon: '',
		mycelium_space_id: '',
		opus_workspace: '',
		agenda_calendar: 0,
		courrier_account: 0,
		created_at: '',
		updated_at: '',
		...extra
	};
}

describe('pickSpace', () => {
	const items = [space('home'), space('work')];

	it('keeps the remembered space while it exists', () => {
		expect(pickSpace(items, 'work')).toBe('work');
	});

	it('falls back to the default space when the remembered one is gone', () => {
		expect(pickSpace(items, 'deleted')).toBe('home');
		expect(pickSpace(items, null)).toBe('home');
		expect(pickSpace([], 'work')).toBeNull();
	});
});

describe('withSpace', () => {
	it('adds space_id only when a space is chosen', () => {
		expect(withSpace('/sessions', null)).toBe('/sessions');
		expect(withSpace('/sessions', 'a b')).toBe('/sessions?space_id=a%20b');
		expect(withSpace('/jobs?x=1', 'w')).toBe('/jobs?x=1&space_id=w');
	});
});

describe('space drafts', () => {
	it('shows unset ids as empty and clears them back to zero', () => {
		const draft = draftFor(space('w', { name: 'Work', courrier_account: 4 }));
		expect(draft).toEqual({ name: 'Work', courrier_account: '4', agenda_calendar: '', opus_workspace: '' });
		draft.agenda_calendar = 'not a number';
		draft.courrier_account = ' 7 ';
		expect(inputFromDraft(draft)).toEqual({ name: 'Work', courrier_account: 7, agenda_calendar: 0, opus_workspace: '' });
	});
});

describe('listJobs in a space', () => {
	const originalFetch = globalThis.fetch;

	beforeEach(() => {
		globalThis.fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => [] });
	});

	afterEach(() => {
		globalThis.fetch = originalFetch;
	});

	it('filters the jobs by the current space', async () => {
		await listJobs('work');
		expect(globalThis.fetch).toHaveBeenCalledWith('/api/jobs?space_id=work', expect.anything());
	});
});
