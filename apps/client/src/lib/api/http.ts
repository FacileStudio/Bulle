import { App } from '@capacitor/app';
import { apiFetch, authedInit, backend } from '$lib/backend';
import { isNative } from '$lib/platform';
import { createSseParser, type SseFrame } from '$lib/sse';
import { withSpace } from '$lib/spaces.svelte';
import {
	RUN_EVENT_TYPES,
	type ChatApi,
	type RunEventType,
	type SeqEvent,
	type StreamHandlers
} from './types';

const MAX_MANUAL_RETRIES = 5;

function post(body?: unknown): RequestInit {
	return body === undefined ? { method: 'POST' } : { method: 'POST', body: JSON.stringify(body) };
}

function openRunStream(runId: string, lastSeq: number, handlers: StreamHandlers) {
	let closed = false;
	let retries = 0;
	let latest = lastSeq;
	let generation = 0;
	let controller: AbortController | null = null;
	let timer: ReturnType<typeof setTimeout> | undefined;

	function schedule(delay: number) {
		timer = setTimeout(connect, delay);
	}

	function fail(mine: number) {
		if (closed || mine !== generation) return;
		retries += 1;
		if (retries > MAX_MANUAL_RETRIES) {
			handlers.onFatal('The live connection dropped and could not be restored.');
			return;
		}
		handlers.onReconnecting();
		schedule(Math.min(1000 * 2 ** retries, 15000));
	}

	function dispatch(frame: SseFrame) {
		const type = frame.event as RunEventType;
		if (!RUN_EVENT_TYPES.includes(type)) return;
		const seq = Number(frame.id);
		if (Number.isFinite(seq) && seq > 0) latest = Math.max(latest, seq);
		let data: unknown;
		try {
			data = JSON.parse(frame.data);
		} catch {
			return;
		}
		handlers.onEvent({ type, data, seq } as SeqEvent);
	}

	async function read(response: Response, mine: number) {
		const parse = createSseParser((frame) => {
			if (!closed && mine === generation) dispatch(frame);
		});
		const decoder = new TextDecoder();
		const reader = response.body!.getReader();
		for (;;) {
			const { done, value } = await reader.read();
			if (done) break;
			parse(decoder.decode(value, { stream: true }));
		}
	}

	async function connect() {
		controller?.abort();
		clearTimeout(timer);
		const mine = ++generation;
		const current = new AbortController();
		controller = current;
		const query = latest > 0 ? `?last_event_id=${latest}` : '';
		try {
			const init = await authedInit({
				headers: { Accept: 'text/event-stream' },
				signal: current.signal
			});
			const response = await fetch(
				`${backend.baseUrl}/runs/${encodeURIComponent(runId)}/events${query}`,
				init
			);
			if (closed || mine !== generation) return;
			if (!response.ok || !response.body) return fail(mine);
			retries = 0;
			handlers.onOpen();
			await read(response, mine);
			if (closed || mine !== generation) return;
			handlers.onReconnecting();
			schedule(1000);
		} catch {
			fail(mine);
		}
	}

	void connect();

	let appListener: Promise<{ remove: () => Promise<void> }> | undefined;
	if (isNative()) {
		appListener = App.addListener('appStateChange', ({ isActive }) => {
			if (isActive && !closed) void connect();
		});
	}

	return {
		close() {
			closed = true;
			clearTimeout(timer);
			controller?.abort();
			void appListener?.then((handle) => handle.remove());
		}
	};
}

export const httpApi: ChatApi = {
	listSessions: (spaceId) => apiFetch(withSpace('/sessions', spaceId)),
	createSession: (title, spaceId) =>
		apiFetch(
			'/sessions',
			post({ ...(title ? { title } : {}), ...(spaceId ? { space_id: spaceId } : {}) })
		),
	getSession: (id) => apiFetch(`/sessions/${encodeURIComponent(id)}`),
	deleteSession: (id) => apiFetch(`/sessions/${encodeURIComponent(id)}`, { method: 'DELETE' }),
	sendMessage: (sessionId, text) =>
		apiFetch(`/sessions/${encodeURIComponent(sessionId)}/messages`, post({ text })),
	stopRun: (runId) => apiFetch(`/runs/${encodeURIComponent(runId)}/stop`, post()),
	decide: (runId, callId, approve) =>
		apiFetch(
			`/runs/${encodeURIComponent(runId)}/approvals/${encodeURIComponent(callId)}`,
			post({ approve })
		),
	openRunStream,
	getComputer: () => apiFetch('/computer'),
	stopComputer: () => apiFetch('/computer/stop', post())
};
