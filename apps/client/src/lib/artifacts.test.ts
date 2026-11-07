import { describe, expect, it } from 'vitest';
import { ApiError } from './backend';
import { artifactCards, artifactError, filterArtifacts, groupByMonth, type Artifact } from './artifacts';
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

describe('artifact lists', () => {
	const items = [
		{ id: '1', title: 'Trip budget', updated: '2026-10-12T10:00:00Z' },
		{ id: '2', title: 'Weekly report', updated: '2026-11-05T10:00:00Z' },
		{ id: '3', title: 'Budget review', updated: '2026-11-03T10:00:00Z' }
	] as Artifact[];

	it('filters by every word of the query', () => {
		expect(filterArtifacts(items, 'budget').map((a) => a.id)).toEqual(['1', '3']);
		expect(filterArtifacts(items, 'review BUDGET').map((a) => a.id)).toEqual(['3']);
		expect(filterArtifacts(items, '  ')).toHaveLength(3);
	});

	it('groups newest first by month', () => {
		const months = groupByMonth(items, 'en');
		expect(months.map((m) => m.items.map((a) => a.id))).toEqual([['2', '3'], ['1']]);
		expect(months[0].label).toBe('November 2026');
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
