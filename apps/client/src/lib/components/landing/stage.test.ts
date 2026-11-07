import { describe, expect, it } from 'vitest';
import { loop, sleep, typeOut } from './stage';

describe('landing stage', () => {
	it('stops the loop when the signal aborts mid-script', async () => {
		const controller = new AbortController();
		let runs = 0;
		const done = loop(async (signal) => {
			runs++;
			if (runs === 2) controller.abort();
			await sleep(5, signal);
		}, controller.signal);
		await expect(done).resolves.toBeUndefined();
		expect(runs).toBe(2);
	});

	it('types text one character per step', async () => {
		const seen: string[] = [];
		await typeOut('abc', (v) => seen.push(v), 0, new AbortController().signal);
		expect(seen).toEqual(['a', 'ab', 'abc']);
	});
});
