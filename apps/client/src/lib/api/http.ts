import { apiFetch, backend } from '$lib/backend';
import { RUN_EVENT_TYPES, type ChatApi, type SeqEvent, type StreamHandlers } from './types';

const MAX_MANUAL_RETRIES = 5;

function post(body?: unknown): RequestInit {
	return body === undefined ? { method: 'POST' } : { method: 'POST', body: JSON.stringify(body) };
}

function openRunStream(runId: string, lastSeq: number, handlers: StreamHandlers) {
	let source: EventSource | null = null;
	let closed = false;
	let retries = 0;
	let latest = lastSeq;
	let timer: ReturnType<typeof setTimeout> | undefined;

	function connect() {
		const query = latest > 0 ? `?last_event_id=${latest}` : '';
		const next = new EventSource(
			`${backend.baseUrl}/runs/${encodeURIComponent(runId)}/events${query}`
		);
		source = next;

		next.onopen = () => {
			retries = 0;
			handlers.onOpen();
		};

		for (const type of RUN_EVENT_TYPES) {
			next.addEventListener(type, (raw) => {
				const message = raw as MessageEvent<string>;
				const seq = Number(message.lastEventId);
				if (Number.isFinite(seq) && seq > 0) latest = Math.max(latest, seq);
				let data: unknown;
				try {
					data = JSON.parse(message.data);
				} catch {
					return;
				}
				handlers.onEvent({ type, data, seq } as SeqEvent);
			});
		}

		next.onerror = () => {
			if (closed) return;
			if (next.readyState === EventSource.CONNECTING) {
				handlers.onReconnecting();
				return;
			}
			next.close();
			retries += 1;
			if (retries > MAX_MANUAL_RETRIES) {
				handlers.onFatal('The live connection dropped and could not be restored.');
				return;
			}
			handlers.onReconnecting();
			timer = setTimeout(connect, Math.min(1000 * 2 ** retries, 15000));
		};
	}

	connect();

	return {
		close() {
			closed = true;
			if (timer) clearTimeout(timer);
			source?.close();
		}
	};
}

export const httpApi: ChatApi = {
	listSessions: () => apiFetch('/sessions'),
	createSession: (title) => apiFetch('/sessions', post(title ? { title } : {})),
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
	startComputer: () => apiFetch('/computer/start', post()),
	stopComputer: () => apiFetch('/computer/stop', post())
};
