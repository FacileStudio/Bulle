import { describe, expect, it } from 'vitest';
import { needsSystemNotification, routeFor } from './push-route';

describe('routeFor', () => {
	it('opens the job conversation and falls back to the chat list', () => {
		expect(routeFor('sessions/abc123')).toBe('/s/abc123');
		expect(routeFor('')).toBe('/sessions');
		expect(routeFor('sessions/../admin')).toBe('/sessions');
		expect(routeFor(undefined)).toBe('/sessions');
	});
});

describe('needsSystemNotification', () => {
	const origin = 'https://bulle.facile.studio';

	it('stays silent while a focused app window shows the live notification', () => {
		const focused = { url: `${origin}/s/abc`, focused: true, visibilityState: 'visible' as const };
		expect(needsSystemNotification([focused], origin)).toBe(false);
		expect(needsSystemNotification([{ ...focused, url: `${origin}/ideas` }], origin)).toBe(false);
	});

	it('notifies when no app window is focused and visible', () => {
		expect(needsSystemNotification([], origin)).toBe(true);
		expect(
			needsSystemNotification([{ url: `${origin}/s/abc`, focused: false, visibilityState: 'visible' }], origin)
		).toBe(true);
		expect(
			needsSystemNotification([{ url: `${origin}/s/abc`, focused: true, visibilityState: 'hidden' }], origin)
		).toBe(true);
		expect(
			needsSystemNotification([{ url: 'https://other.example/s/abc', focused: true }], origin)
		).toBe(true);
	});
});
