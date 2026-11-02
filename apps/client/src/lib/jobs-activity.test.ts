import { describe, expect, it } from 'vitest';
import type { Job } from './api';
import { anyRunning } from './jobs-activity.svelte';

const job = (running?: boolean) => ({ id: 'j', running }) as Job;

describe('anyRunning', () => {
	it('is true only when a job reports a run in progress', () => {
		expect(anyRunning([])).toBe(false);
		expect(anyRunning([job(false), job()])).toBe(false);
		expect(anyRunning([job(false), job(true)])).toBe(true);
	});
});
