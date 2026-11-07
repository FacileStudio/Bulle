import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SEARCH_DELAY_MS, debounce, searchSessions, shouldSearch } from './search';

describe('debounce', () => {
	beforeEach(() => vi.useFakeTimers());
	afterEach(() => vi.useRealTimers());

	it('runs once with the last arguments after the pause', () => {
		const fn = vi.fn();
		const search = debounce(fn, SEARCH_DELAY_MS);
		search.call('to');
		vi.advanceTimersByTime(200);
		search.call('tok');
		vi.advanceTimersByTime(249);
		expect(fn).not.toHaveBeenCalled();
		vi.advanceTimersByTime(1);
		expect(fn).toHaveBeenCalledTimes(1);
		expect(fn).toHaveBeenCalledWith('tok');
	});

	it('drops a pending call on cancel', () => {
		const fn = vi.fn();
		const search = debounce(fn, SEARCH_DELAY_MS);
		search.call('to');
		search.cancel();
		vi.advanceTimersByTime(SEARCH_DELAY_MS * 2);
		expect(fn).not.toHaveBeenCalled();
	});
});

describe('session search', () => {
	afterEach(() => vi.unstubAllGlobals());

	it('needs two characters', () => {
		expect(shouldSearch(' a ')).toBe(false);
		expect(shouldSearch('ab')).toBe(true);
	});

	it('calls the search route with the space and tolerates null results', async () => {
		const fetch = vi.fn(async () => new Response(JSON.stringify({ results: null }), { status: 200 }));
		vi.stubGlobal('fetch', fetch);
		expect(await searchSessions(' tax & co ', 's1')).toEqual([]);
		expect(fetch).toHaveBeenCalledWith('/api/sessions/search?q=tax%20%26%20co&space_id=s1', expect.anything());
	});
});
