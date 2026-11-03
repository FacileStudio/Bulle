import { describe, expect, it } from 'vitest';
import { noteAction } from './notify';

describe('noteAction', () => {
	const route = '/s/abc';

	it('stays quiet for the conversation the user is looking at', () => {
		expect(noteAction(route, { openRoute: route, visible: true, focused: true })).toBe('skip');
		expect(noteAction(route, { openRoute: route, visible: true, focused: false })).toBe('skip');
	});

	it('notifies for that conversation once the tab is hidden', () => {
		expect(noteAction(route, { openRoute: route, visible: false, focused: false })).toBe('system');
	});

	it('uses a toast in a focused tab showing something else', () => {
		expect(noteAction(route, { openRoute: '/s/other', visible: true, focused: true })).toBe('toast');
		expect(noteAction(route, { openRoute: null, visible: true, focused: true })).toBe('toast');
	});

	it('uses a system notification when the tab has no focus', () => {
		expect(noteAction(route, { openRoute: '/s/other', visible: true, focused: false })).toBe('system');
	});
});
