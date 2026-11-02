import { describe, expect, it } from 'vitest';
import { getCaretCoordinates } from './caret';

describe('getCaretCoordinates', () => {
	it('returns zero coordinates in non-browser environment', () => {
		const coords = getCaretCoordinates(null as unknown as HTMLTextAreaElement);
		expect(coords).toEqual({ x: 0, y: 0 });
	});

	it('computes relative coordinates from element box', () => {
		const element = {
			value: 'hello world',
			selectionStart: 5,
			clientWidth: 200,
			scrollLeft: 0,
			scrollTop: 0,
			getBoundingClientRect: () => ({ left: 50, top: 100, width: 200, height: 40 })
		} as unknown as HTMLTextAreaElement;

		const coords = getCaretCoordinates(element, 5);
		expect(coords).toHaveProperty('x');
		expect(coords).toHaveProperty('y');
	});
});
