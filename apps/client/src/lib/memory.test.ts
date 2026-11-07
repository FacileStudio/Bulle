import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'svelte/server';

const apiFetch = vi.fn();

vi.mock('$lib/backend', async (importOriginal) => ({
	...(await importOriginal<typeof import('./backend')>()),
	apiFetch: (...args: unknown[]) => apiFetch(...args)
}));

const { ApiError } = await import('./backend');
const memory = await import('./memory');
const { default: MemorySettings } = await import('./components/settings/MemorySettings.svelte');

beforeEach(() => apiFetch.mockReset());

describe('memoryPagePath', () => {
	it('escapes each segment and keeps folders', () => {
		expect(memory.memoryPagePath('people/ann lee.md', 's 1')).toBe('/memory/pages/people/ann%20lee.md?space_id=s%201');
		expect(memory.memoryPagePath('50%.md', null)).toBe('/memory/pages/50%25.md');
	});
});

describe('memory requests', () => {
	it('lists pages and treats a missing flag as unavailable', async () => {
		apiFetch.mockResolvedValueOnce({});
		expect(await memory.listMemoryPages('s1')).toEqual({ pages: [], available: false });
		expect(apiFetch).toHaveBeenCalledWith('/memory/pages?space_id=s1');
	});

	it('saves with PUT and deletes with DELETE on the page path', async () => {
		apiFetch.mockResolvedValue({ path: 'a.md', content: 'x' });
		await memory.saveMemoryPage('a.md', 'x', null);
		await memory.deleteMemoryPage('a.md', null);
		expect(apiFetch).toHaveBeenNthCalledWith(1, '/memory/pages/a.md', { method: 'PUT', body: '{"content":"x"}' });
		expect(apiFetch).toHaveBeenNthCalledWith(2, '/memory/pages/a.md', { method: 'DELETE' });
	});
});

describe('pageLabel and memoryError', () => {
	it('names the profile and strips .md elsewhere', () => {
		expect(memory.pageLabel('profile.md')).not.toBe('profile.md');
		expect(memory.pageLabel('people/ann.md')).toBe('people/ann');
	});

	it('asks to connect Mycelium on 412 only', () => {
		expect(memory.memoryError(new ApiError('no', 412), 'fallback').connect).toBe(true);
		expect(memory.memoryError(new ApiError('gone', 404), 'fallback').connect).toBe(false);
		expect(memory.memoryError(new Error('boom'), 'fallback').text).toBe('fallback');
	});
});

describe('MemorySettings', () => {
	it('renders the section heading with a loading list and no borders', () => {
		const html = render(MemorySettings).body;
		expect(html).toContain('id="memory-title"');
		expect(html).toContain('aria-busy="true"');
		expect(html).not.toMatch(/class="[^"]*\bborder(?!-0)/);
	});
});
