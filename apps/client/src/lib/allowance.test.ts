import { describe, expect, it } from 'vitest';
import { allowance } from './allowance';

const budget = (spent_usd: number, limit_usd = 5, exempt = false) => ({
	limit_usd,
	spent_usd,
	resets_at: '2026-12-01T00:00:00Z',
	exempt
});

describe('allowance', () => {
	it('stays quiet under 80 percent', () => {
		expect(allowance(budget(3.99))).toEqual({ percent: 80, tone: 'ok' });
	});

	it('warns from 80 percent', () => {
		expect(allowance(budget(4))).toEqual({ percent: 80, tone: 'warn' });
	});

	it('is over at the limit and caps the bar at 100 percent', () => {
		expect(allowance(budget(5))).toEqual({ percent: 100, tone: 'over' });
		expect(allowance(budget(12))).toEqual({ percent: 100, tone: 'over' });
	});

	it('shows no meter without a budget, for an exempt user or before the server answers', () => {
		expect(allowance(budget(1, 0))).toBeNull();
		expect(allowance(budget(9, 5, true))).toBeNull();
		expect(allowance(undefined)).toBeNull();
	});
});
