import { afterEach, describe, expect, it } from 'vitest';
import { setApi } from '$lib/api';
import type { ChatApi, SeqEvent, StreamHandlers } from '$lib/api/types';
import { httpApi } from '$lib/api/http';
import { ChatController } from './chat.svelte';

function fakeApi() {
	const sent: string[] = [];
	const opened: { runId: string; lastSeq: number; handlers: StreamHandlers }[] = [];
	const api = {
		getSession: async () => ({
			session: { id: 's1', title: 'Hello', updated_at: '2026-01-01T00:00:00Z' },
			messages: [{ seq: 1, role: 'user', content: 'earlier' }],
			active_run_id: 'run-active'
		}),
		sendMessage: async (_id: string, text: string) => {
			sent.push(text);
			return { run_id: 'run-new' };
		},
		listSessions: async () => [],
		openRunStream: (runId: string, lastSeq: number, handlers: StreamHandlers) => {
			opened.push({ runId, lastSeq, handlers });
			return { close() {} };
		}
	} as unknown as ChatApi;
	return { api, sent, opened };
}

afterEach(() => setApi(httpApi));

describe('ChatController', () => {
	it('resumes the active run on open and folds streamed events into state', async () => {
		const fake = fakeApi();
		setApi(fake.api);
		const chat = new ChatController();
		await chat.open('s1');
		expect(chat.title).toBe('Hello');
		expect(fake.opened[0].runId).toBe('run-active');
		expect(chat.state.running).toBe(true);

		const { handlers } = fake.opened[0];
		handlers.onOpen();
		expect(chat.connection).toBe('live');
		handlers.onEvent({ type: 'TEXT_MESSAGE_CONTENT', data: { delta: 'yo' }, seq: 1 } as SeqEvent);
		handlers.onEvent({ type: 'RUN_FINISHED', data: { cost_usd: 0.1 }, seq: 2 } as SeqEvent);
		expect(chat.state.items.at(-1)?.text).toBe('yo');
		expect(chat.state.running).toBe(false);
		expect(chat.connection).toBe('idle');
	});

	it('refuses to send while a run is active and sends otherwise', async () => {
		const fake = fakeApi();
		setApi(fake.api);
		const chat = new ChatController();
		await chat.open('s1');
		expect(await chat.send('nope')).toBe(false);
		fake.opened[0].handlers.onEvent({
			type: 'RUN_FINISHED',
			data: {},
			seq: 1
		} as SeqEvent);
		expect(await chat.send('  next  ')).toBe(true);
		expect(fake.sent).toEqual(['next']);
		expect(fake.opened[1].runId).toBe('run-new');
	});
});
