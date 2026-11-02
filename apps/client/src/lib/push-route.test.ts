import { describe, expect, it } from 'vitest';
import { routeFor } from './push-route';

describe('routeFor', () => {
	it('opens the job conversation and falls back to the chat list', () => {
		expect(routeFor('sessions/abc123')).toBe('/s/abc123');
		expect(routeFor('')).toBe('/sessions');
		expect(routeFor('sessions/../admin')).toBe('/sessions');
		expect(routeFor(undefined)).toBe('/sessions');
	});
});
