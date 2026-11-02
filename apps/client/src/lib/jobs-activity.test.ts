import { describe, expect, it } from 'vitest';
import type { Job } from './api';
import { anyRunning, runSessions } from './jobs-activity.svelte';

const job = (running?: boolean) => ({ id: 'j', running }) as Job;

describe('anyRunning', () => {
	it('is true only when a job reports a run in progress', () => {
		expect(anyRunning([])).toBe(false);
		expect(anyRunning([job(false), job()])).toBe(false);
		expect(anyRunning([job(false), job(true)])).toBe(true);
	});
});

describe('runSessions', () => {
	it('changes when a job starts a run in a new conversation', () => {
		const a = [{ id: 'a', last_session_id: 's1' }, { id: 'b' }] as Job[];
		const b = [{ id: 'a', last_session_id: 's2' }, { id: 'b' }] as Job[];
		expect(runSessions(a)).toBe('s1');
		expect(runSessions(a)).not.toBe(runSessions(b));
	});
});
