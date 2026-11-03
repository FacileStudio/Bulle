import { beforeAll, describe, expect, it } from 'vitest';
import { desktopSidebar, isSidebarShortcut, setCollapsed } from './mobile-sidebar.svelte';

const key = (over: Partial<KeyboardEvent>) => ({ key: 'b', ctrlKey: false, metaKey: false, altKey: false, shiftKey: false, ...over });

const store = new Map<string, string>();

beforeAll(() => {
	Object.assign(globalThis, {
		localStorage: {
			getItem: (k: string) => store.get(k) ?? null,
			setItem: (k: string, v: string) => void store.set(k, v)
		}
	});
});

describe('docked sidebar', () => {
	it('toggles on Ctrl+B and Cmd+B only', () => {
		expect(isSidebarShortcut(key({ ctrlKey: true }))).toBe(true);
		expect(isSidebarShortcut(key({ metaKey: true, key: 'B' }))).toBe(true);
		expect(isSidebarShortcut(key({}))).toBe(false);
		expect(isSidebarShortcut(key({ ctrlKey: true, shiftKey: true }))).toBe(false);
		expect(isSidebarShortcut(key({ ctrlKey: true, key: 'k' }))).toBe(false);
	});

	it('remembers the collapsed state', () => {
		setCollapsed(true);
		expect(desktopSidebar.collapsed).toBe(true);
		expect(store.get('bulle.sidebar.collapsed')).toBe('1');
		setCollapsed(false);
		expect(store.get('bulle.sidebar.collapsed')).toBe('0');
	});
});
