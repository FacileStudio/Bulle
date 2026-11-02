import { describe, expect, it } from 'vitest';
import type { RunEvent, SeqEvent } from '$lib/api/types';
import {
	addUser,
	applyEvent,
	beginRun,
	failRun,
	fromHistory,
	initialState,
	setApproval,
	totalTokens,
	type ChatState
} from './reducer';
import { approvalFields, approvalIsDestructive, approvalLabel, groupTools, summarizeArgs, toolKind } from './tools';

function feed(state: ChatState, events: RunEvent[]): ChatState {
	return events.reduce(
		(acc, event, index) => applyEvent(acc, { ...event, seq: index + 1 } as SeqEvent),
		state
	);
}

const live = (state: ChatState) => state.items[state.items.length - 1];

describe('stream reducer', () => {
	it('keeps only the text after the last tool call as the answer', () => {
		const state = feed(addUser(initialState(), 'brief'), [
			{ type: 'RUN_STARTED', data: { run_id: 'r1' } },
			{ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'Let me check.' } },
			{ type: 'TOOL_CALL_START', data: { call_id: 'c1', name: 'courrier_inbox', args: {} } },
			{ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'Two emails need you.' } }
		] as RunEvent[]);
		expect(live(state).text).toBe('Let me check.Two emails need you.');
		expect(live(state).answer).toBe('Two emails need you.');
	});

	it('accumulates text and thinking into one streaming assistant item', () => {
		const state = feed(addUser(initialState(), 'hi'), [
			{ type: 'RUN_STARTED', data: { run_id: 'r1' } },
			{ type: 'THINKING', data: { delta: 'hmm ' } },
			{ type: 'THINKING', data: { delta: 'ok' } },
			{ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'Hel' } },
			{ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'lo' } }
		]);
		expect(state.items).toHaveLength(2);
		expect(live(state)).toMatchObject({
			role: 'assistant',
			text: 'Hello',
			thinking: 'hmm ok',
			streaming: true
		});
		expect(state.running).toBe(true);
		expect(state.runId).toBe('r1');
	});

	it('tracks a tool call from running to ok with duration', () => {
		const state = feed(initialState(), [
			{ type: 'RUN_STARTED', data: { run_id: 'r1' } },
			{ type: 'TOOL_CALL_START', data: { call_id: 'c1', name: 'bash', args: { command: 'ls' } } },
			{
				type: 'TOOL_CALL_END',
				data: { call_id: 'c1', ok: true, output_preview: 'a b', duration_ms: 42 }
			},
			{ type: 'TOOL_CALL_START', data: { call_id: 'c2', name: 'read_file', args: {} } },
			{ type: 'TOOL_CALL_END', data: { call_id: 'c2', ok: false } }
		]);
		const tools = live(state).tools;
		expect(tools[0]).toMatchObject({ status: 'ok', durationMs: 42, outputPreview: 'a b' });
		expect(tools[1].status).toBe('failed');
	});

	it('opens an approval on INTERRUPT and settles it when the call ends', () => {
		let state = feed(initialState(), [
			{ type: 'RUN_STARTED', data: { run_id: 'r1' } },
			{ type: 'TOOL_CALL_START', data: { call_id: 'c1', name: 'bash', args: 'rm x' } },
			{ type: 'INTERRUPT', data: { call_id: 'c1', tool: 'bash', args: 'rm x' } }
		]);
		expect(live(state).approvals[0].state).toBe('pending');
		state = setApproval(state, 'c1', 'sending');
		expect(live(state).approvals[0].state).toBe('sending');
		state = applyEvent(state, {
			type: 'TOOL_CALL_END',
			data: { call_id: 'c1', ok: true },
			seq: 10
		});
		expect(live(state).approvals[0].state).toBe('approved');
	});

	it('finishes a run and accumulates tokens and cost', () => {
		const state = feed(initialState(), [
			{ type: 'RUN_STARTED', data: { run_id: 'r1' } },
			{ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'done' } },
			{
				type: 'RUN_FINISHED',
				data: { stop: 'end_turn', usage: { input_tokens: 100, output_tokens: 20 }, cost_usd: 0.5 }
			}
		]);
		expect(state.running).toBe(false);
		expect(live(state).streaming).toBe(false);
		expect(state.tokens).toBe(120);
		expect(state.costUsd).toBe(0.5);
		expect(state.stop).toBe('end_turn');
	});

	it('records RUN_ERROR on the assistant item and the banner', () => {
		const state = feed(initialState(), [
			{ type: 'RUN_STARTED', data: { run_id: 'r1' } },
			{ type: 'RUN_ERROR', data: { message: 'boom' } }
		]);
		expect(state.running).toBe(false);
		expect(state.error).toBe('boom');
		expect(live(state).error).toBe('boom');
	});

	it('ignores events already seen, so a replay after reconnect is idempotent', () => {
		const events: RunEvent[] = [
			{ type: 'RUN_STARTED', data: { run_id: 'r1' } },
			{ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'a' } },
			{ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'b' } }
		];
		const once = feed(initialState(), events);
		const twice = events.reduce(
			(acc, event, index) => applyEvent(acc, { ...event, seq: index + 1 } as SeqEvent),
			once
		);
		expect(live(twice).text).toBe('ab');
		expect(twice.lastSeq).toBe(3);
	});

	it('resumes onto history without duplicating the user turn', () => {
		const base = fromHistory([
			{ seq: 1, role: 'user', content: 'q' },
			{ seq: 2, role: 'tool', content: 'ignored' },
			{ seq: 3, role: 'assistant', content: 'a' }
		]);
		expect(base.items.map((i) => i.text)).toEqual(['q', 'a']);
		const state = feed(beginRun(base, 'r9'), [
			{ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'more' } }
		]);
		expect(state.items).toHaveLength(3);
		expect(live(state).text).toBe('more');
	});

	it('failRun settles a streaming turn', () => {
		const state = failRun(beginRun(initialState(), 'r1'), 'lost');
		expect(state.running).toBe(false);
		expect(live(state).streaming).toBe(false);
	});

	it('restores thinking and tool calls with status and output preview from history', () => {
		const state = fromHistory([
			{ seq: 1, role: 'user', content: 'check files' },
			{
				seq: 2,
				role: 'assistant',
				content: 'done',
				thinking: 'searching...',
				tools: [
					{
						call_id: 'c1',
						name: 'bash',
						args: { command: 'ls' },
						status: 'ok',
						output_preview: 'file.txt',
						duration_ms: 123
					},
					{
						call_id: 'c2',
						name: 'read_file',
						args: { path: 'file.txt' },
						status: 'failed',
						output_preview: 'permission denied'
					}
				]
			}
		]);
		expect(state.items).toHaveLength(2);
		expect(state.items[0]).toMatchObject({
			role: 'user',
			text: 'check files',
			thinking: '',
			tools: []
		});
		expect(state.items[1]).toMatchObject({
			role: 'assistant',
			text: 'done',
			thinking: 'searching...',
			tools: [
				{
					callId: 'c1',
					name: 'bash',
					args: { command: 'ls' },
					status: 'ok',
					outputPreview: 'file.txt',
					durationMs: 123
				},
				{
					callId: 'c2',
					name: 'read_file',
					args: { path: 'file.txt' },
					status: 'failed',
					outputPreview: 'permission denied'
				}
			]
		});
	});
});

describe('helpers', () => {
	it('sums usage', () => {
		expect(totalTokens({ total_tokens: 7, input_tokens: 1 })).toBe(7);
		expect(totalTokens({ input_tokens: 2, output_tokens: 3 })).toBe(5);
		expect(totalTokens(undefined)).toBe(0);
	});

	it('groups tools by kind in first-seen order', () => {
		const call = (callId: string, name: string) => ({
			callId,
			name,
			args: {},
			status: 'ok' as const
		});
		const groups = groupTools([call('1', 'read_file'), call('2', 'bash'), call('3', 'edit_file')]);
		expect(groups.map((g) => [g.kind, g.calls.length])).toEqual([
			['files', 2],
			['commands', 1]
		]);
		expect(toolKind('web_fetch')).toBe('web');
		expect(toolKind('mystery')).toBe('other');
		expect(toolKind('courrier_read')).toBe('apps');
	});

	it('describes approvals in plain words', () => {
		expect(approvalLabel('courrier_send')).toBe('Send an email');
		expect(approvalLabel('bash')).toBe('bash');
		expect(approvalFields({ to: ['a@x', 'b@x'], subject: 'Hi', cc: '' })).toEqual([
			{ key: 'to', value: 'a@x, b@x' },
			{ key: 'subject', value: 'Hi' }
		]);
		expect(approvalFields({ nested: { a: 1 } })).toBeNull();
		expect(approvalIsDestructive('courrier_update', { action: 'delete' })).toBe(true);
		expect(approvalIsDestructive('agenda_delete_event', {})).toBe(true);
		expect(approvalIsDestructive('courrier_send', {})).toBe(false);
	});

	it('summarizes args', () => {
		expect(summarizeArgs({ command: 'ls -la' })).toBe('ls -la');
		expect(summarizeArgs({ x: 1 })).toBe('{"x":1}');
		expect(summarizeArgs('a'.repeat(200), 10)).toHaveLength(10);
	});
});
