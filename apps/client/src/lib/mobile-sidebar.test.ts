import { describe, expect, it } from 'vitest';
import { swipeIntent } from './mobile-sidebar.svelte';

describe('swipeIntent', () => {
	it('opens on a right swipe and closes on a left one', () => {
		expect(swipeIntent(80, 10, false)).toBe('open');
		expect(swipeIntent(-80, 10, true)).toBe('close');
	});

	it('ignores short, vertical and redundant swipes', () => {
		expect(swipeIntent(40, 0, false)).toBeNull();
		expect(swipeIntent(80, 70, false)).toBeNull();
		expect(swipeIntent(80, 0, true)).toBeNull();
		expect(swipeIntent(-80, 0, false)).toBeNull();
	});
});
