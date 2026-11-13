import { describe, expect, it } from 'vitest';
import type { SessionSummary } from '$lib/api';
import { applyLiveEvent, parseLiveFrame, readAfter, type LiveEvent, type LiveSession } from './events';

const row = (id: string, updated: string, extra: Partial<SessionSummary> = {}): SessionSummary => ({
	id,
	space_id: 'sp1',
	title: id,
	updated_at: updated,
	status: 'idle',
	...extra
});

const data = (id: string, updated: string, space = 'sp1'): LiveSession => ({
	session_id: id,
	space_id: space,
	title: `${id} renamed`,
	updated_at: updated,
	status: 'running'
});

const session = (id: string, updated: string, space = 'sp1'): LiveEvent => ({
	type: 'session.updated',
	data: data(id, updated, space)
});

describe('applyLiveEvent', () => {
	const items = [row('a', '2026-11-03T10:00:00Z'), row('b', '2026-11-03T09:00:00Z')];

	it('adds a conversation created on another device at the top', () => {
		const created: LiveEvent = { type: 'session.created', data: data('c', '2026-11-03T11:00:00Z') };
		const next = applyLiveEvent(items, created, 'sp1');
		expect(next.map((s) => s.id)).toEqual(['c', 'a', 'b']);
		expect(next[0].title).toBe('c renamed');
	});

	it('renames and reorders an existing conversation', () => {
		const next = applyLiveEvent(items, session('b', '2026-11-03T12:00:00Z'), 'sp1');
		expect(next.map((s) => s.id)).toEqual(['b', 'a']);
		expect(next[0]).toMatchObject({ title: 'b renamed', status: 'running' });
	});

	it('ignores and drops conversations of another space', () => {
		expect(applyLiveEvent(items, session('c', '2026-11-03T12:00:00Z', 'sp2'), 'sp1')).toHaveLength(2);
		expect(applyLiveEvent(items, session('a', '2026-11-03T12:00:00Z', 'sp2'), 'sp1').map((s) => s.id)).toEqual(['b']);
		expect(applyLiveEvent(items, session('c', '2026-11-03T12:00:00Z', 'sp2'), null)).toHaveLength(3);
	});

	it('removes a deleted conversation', () => {
		const next = applyLiveEvent(items, { type: 'session.deleted', data: { session_id: 'a' } }, 'sp1');
		expect(next.map((s) => s.id)).toEqual(['b']);
		expect(applyLiveEvent(items, { type: 'session.deleted', data: { session_id: 'zz' } }, 'sp1')).toBe(items);
	});

	it('follows the status through a run and its approval', () => {
		const run = { session_id: 'a', run_id: 'r1' };
		let next = applyLiveEvent(items, { type: 'run.started', data: { ...run, session_status: 'running' } }, 'sp1');
		expect(next[0].status).toBe('running');
		next = applyLiveEvent(next, { type: 'approval.pending', data: { ...run, call_id: 'c1', session_status: 'waiting' } }, 'sp1');
		expect(next[0].status).toBe('waiting');
		next = applyLiveEvent(next, { type: 'approval.decided', data: { ...run, call_id: 'c1', approved: true, session_status: 'running' } }, 'sp1');
		expect(next[0].status).toBe('running');
		next = applyLiveEvent(next, { type: 'run.finished', data: { ...run, status: 'finished', session_status: 'idle' } }, 'sp1');
		expect(next[0].status).toBe('idle');
		expect(next[1]).toBe(items[1]);
	});
});

describe('unread conversations', () => {
	const job: LiveEvent = {
		type: 'session.created',
		data: { ...data('job', '2026-11-03T11:00:00Z'), channel: 'cron', unread: true }
	};

	it('shows a job conversation unread until it is read on any device', () => {
		const items = [row('a', '2026-11-03T10:00:00Z')];
		const listed = applyLiveEvent(items, job, 'sp1');
		expect(listed[0]).toMatchObject({ id: 'job', unread: true });
		expect(listed[1].unread).toBeUndefined();
		const read = applyLiveEvent(listed, { type: 'session.read', data: { session_id: 'job' } }, 'sp1');
		expect(read[0].unread).toBe(false);
		expect(applyLiveEvent(read, { type: 'session.read', data: { session_id: 'job' } }, 'sp1')).toBe(read);
	});

	it('marks read only the open conversation on a visible screen', () => {
		expect(readAfter(job, 'job', true)).toBe('job');
		expect(readAfter(job, 'job', false)).toBeNull();
		expect(readAfter(job, 'other', true)).toBeNull();
		expect(readAfter(session('job', '2026-11-03T12:00:00Z'), 'job', true)).toBeNull();
	});

	it('reads session.read frames', () => {
		expect(parseLiveFrame({ id: '', event: 'session.read', data: '{"session_id":"a"}' })).toEqual({
			type: 'session.read',
			data: { session_id: 'a' }
		});
	});
});

describe('parseLiveFrame', () => {
	it('reads known events and skips the rest', () => {
		expect(parseLiveFrame({ id: '', event: 'session.deleted', data: '{"session_id":"a"}' })).toEqual({
			type: 'session.deleted',
			data: { session_id: 'a' }
		});
		expect(parseLiveFrame({ id: '', event: 'mystery', data: '{}' })).toBeNull();
		expect(parseLiveFrame({ id: '', event: 'run.started', data: 'not json' })).toBeNull();
	});
});
