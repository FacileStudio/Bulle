import { beforeEach, describe, expect, it, vi } from 'vitest';

const apiFetch = vi.hoisted(() => vi.fn());

vi.mock('$lib/backend', async (importOriginal) => ({
	...(await importOriginal<typeof import('./backend')>()),
	apiFetch: (...args: unknown[]) => apiFetch(...args)
}));

import { ApiError } from './backend';
import {
	artifactCards,
	autoOpenTarget,
	artifactError,
	artifactHref,
	artifactPath,
	artifactSpaces,
	inSpace,
	isEditableTarget,
	normaliseQuery,
	searchArtifacts,
	snippetParts,
	type Artifact
} from './artifacts';

beforeEach(() => apiFetch.mockReset());
import type { SeqEvent } from './api/types';
import type { ToolCall } from './chat/reducer';

function call(name: string, output: string | undefined, status: ToolCall['status'] = 'ok'): ToolCall {
	return { callId: `${name}-${Math.random()}`, name, args: {}, status, outputPreview: output };
}

const out = (id: string, title: string) => JSON.stringify({ artifact: { id, title, space_id: 's1' } });

describe('artifactCards', () => {
	it('reads one card per finished artifact write, latest title wins', () => {
		const cards = artifactCards([
			call('artifact_create', out('a', 'Draft')),
			call('memory_write', out('x', 'not an artifact')),
			call('artifact_create', out('b', 'Other')),
			call('artifact_update', out('a', 'Final'))
		]);
		expect(cards).toEqual([
			{ id: 'a', title: 'Final', space_id: 's1' },
			{ id: 'b', title: 'Other', space_id: 's1' }
		]);
	});

	it('skips running, failed and unreadable calls', () => {
		expect(
			artifactCards([
				call('artifact_create', out('a', 'A'), 'running'),
				call('artifact_create', out('b', 'B'), 'failed'),
				call('artifact_create', 'error: mycelium is down'),
				call('artifact_create', JSON.stringify({ artifact: { title: 'no id' } })),
				call('artifact_update', undefined)
			])
		).toEqual([]);
	});

	it('falls back to a title when the tool sent none', () => {
		expect(artifactCards([call('artifact_create', JSON.stringify({ artifact: { id: 'z' } }))])[0]).toEqual({
			id: 'z',
			title: 'Untitled artifact',
			space_id: ''
		});
	});
});

describe('artifact search', () => {
	it('sends the trimmed query and the cursor, and reads a missing flag as unavailable', async () => {
		apiFetch.mockResolvedValueOnce({ artifacts: [{ id: 'a' }], next_cursor: 'c2', available: true });
		expect(await searchArtifacts('  tea & cake ', 'c1')).toEqual({ artifacts: [{ id: 'a' }], next_cursor: 'c2', available: true });
		expect(apiFetch).toHaveBeenCalledWith('/artifacts/search?q=tea+%26+cake&cursor=c1');
		apiFetch.mockResolvedValueOnce({});
		expect(await searchArtifacts('tea')).toEqual({ artifacts: [], next_cursor: '', available: false });
		expect(apiFetch).toHaveBeenLastCalledWith('/artifacts/search?q=tea');
	});

	it('cuts a query to what the route accepts', () => {
		expect(normaliseQuery(` ${'a'.repeat(250)} `)).toHaveLength(200);
	});

	it('marks every query word in a snippet, ignoring case and regex characters', () => {
		expect(snippetParts('Oolong tea, then more TEA (c++)', 'tea c++')).toEqual([
			{ text: 'Oolong ', hit: false },
			{ text: 'tea', hit: true },
			{ text: ', then more ', hit: false },
			{ text: 'TEA', hit: true },
			{ text: ' (', hit: false },
			{ text: 'c++', hit: true },
			{ text: ')', hit: false }
		]);
		expect(snippetParts('', 'tea')).toEqual([]);
		expect(snippetParts('plain', ' ')).toEqual([{ text: 'plain', hit: false }]);
	});

	it('leaves the slash key to fields that take text', () => {
		expect(isEditableTarget({ tagName: 'INPUT' } as unknown as EventTarget)).toBe(true);
		expect(isEditableTarget({ tagName: 'DIV', isContentEditable: true } as unknown as EventTarget)).toBe(true);
		expect(isEditableTarget({ tagName: 'BUTTON', isContentEditable: false } as unknown as EventTarget)).toBe(false);
		expect(isEditableTarget(null)).toBe(false);
	});
});

describe('artifactError', () => {
	it('names the missing connection, a gone artifact and an outage', () => {
		expect(artifactError(new ApiError('precondition', 412)).connect).toBe(true);
		expect(artifactError(new ApiError('nope', 404)).text).toMatch(/does not exist/);
		expect(artifactError(new ApiError('down', 503)).text).toMatch(/not answering/);
		expect(artifactError(new Error('boom')).text).toBe('Could not load the artifact.');
	});
});

describe('artifact spaces', () => {
	const at = (id: string, mycelium_space: string, space_name = ''): Artifact => ({
		id,
		title: id,
		machine: 'ruche',
		created: '2026-11-10T09:00:00Z',
		updated: '2026-11-10T09:00:00Z',
		mycelium_space,
		space_name
	});
	const items = [at('a', 'team', 'Zeta'), at('b', 'common'), at('c', 'shared', 'Alpha'), at('d', 'team', 'Zeta')];

	it('lists each space once, personal first, then by name', () => {
		expect(artifactSpaces(items).map((s) => s.key)).toEqual(['common', 'shared', 'team']);
	});

	it('filters by space, and keeps everything without one', () => {
		expect(inSpace(items, 'team').map((a) => a.id)).toEqual(['a', 'd']);
		expect(inSpace(items, '')).toHaveLength(4);
	});

	it('opens an artifact in the tree the list named', () => {
		expect(artifactHref(at('x y', 'team'))).toBe('/artifacts/x%20y?in=team');
		expect(artifactPath('x', { myceliumSpace: 'common', spaceId: 's1' })).toBe('/artifacts/x?space_id=s1&mycelium_space=common');
		expect(artifactPath('x', {})).toBe('/artifacts/x');
	});
});

describe('autoOpenTarget', () => {
	const out = (id: string) => JSON.stringify({ artifact: { id, title: id, space_id: 's' } });
	const tool = (callId: string, name = 'artifact_create'): ToolCall => ({ callId, name, args: {}, status: 'ok' });
	const end = (seq: number, callId: string, id: string, ok = true): SeqEvent => ({
		seq,
		type: 'TOOL_CALL_END',
		data: { call_id: callId, ok, output_preview: ok ? out(id) : 'boom' }
	});

	it('opens on a live successful write', () => {
		expect(autoOpenTarget([tool('a')], [end(3, 'a', 'x')], 2, new Set())?.id).toBe('x');
	});

	it('ignores replayed events', () => {
		expect(autoOpenTarget([tool('a')], [end(3, 'a', 'x')], 3, new Set())).toBeNull();
	});

	it('ignores a failed write and other tools', () => {
		expect(autoOpenTarget([tool('a')], [end(3, 'a', 'x', false)], 0, new Set())).toBeNull();
		expect(autoOpenTarget([tool('a', 'memory_write')], [end(3, 'a', 'x')], 0, new Set())).toBeNull();
	});

	it('opens the latest of several writes', () => {
		const events = [end(3, 'a', 'x'), end(4, 'b', 'y')];
		expect(autoOpenTarget([tool('a'), tool('b', 'artifact_update')], events, 0, new Set())?.id).toBe('y');
	});

	it('keeps a closed artifact closed but opens a new one', () => {
		const tools = [tool('a', 'artifact_update'), tool('b')];
		expect(autoOpenTarget(tools, [end(3, 'a', 'x')], 0, new Set(['x']))).toBeNull();
		expect(autoOpenTarget(tools, [end(3, 'a', 'x'), end(4, 'b', 'z')], 0, new Set(['x']))?.id).toBe('z');
	});
});
