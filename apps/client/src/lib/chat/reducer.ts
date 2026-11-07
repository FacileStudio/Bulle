import type { SeqEvent, StoredMessage, Usage } from '$lib/api/types';
import type { Attachment } from './attachments';

export type ToolStatus = 'running' | 'ok' | 'failed';

export type ToolCall = {
	callId: string;
	name: string;
	args: unknown;
	status: ToolStatus;
	outputPreview?: string;
	durationMs?: number;
};

export type ApprovalState = 'pending' | 'sending' | 'approved' | 'denied';

export type Approval = {
	callId: string;
	tool: string;
	args: unknown;
	state: ApprovalState;
	grantable?: boolean;
};

export type ChatItem = {
	id: string;
	role: 'user' | 'assistant';
	text: string;
	answer: string;
	thinking: string;
	tools: ToolCall[];
	approvals: Approval[];
	attachments: Attachment[];
	streaming: boolean;
	error?: string;
};

export type ChatState = {
	items: ChatItem[];
	runId: string | null;
	running: boolean;
	lastSeq: number;
	error: string | null;
	tokens: number;
	costUsd: number;
	stop: string | null;
};

export function initialState(): ChatState {
	return {
		items: [],
		runId: null,
		running: false,
		lastSeq: 0,
		error: null,
		tokens: 0,
		costUsd: 0,
		stop: null
	};
}

function blankItem(id: string, role: ChatItem['role'], text = ''): ChatItem {
	return {
		id,
		role,
		text,
		answer: text,
		thinking: '',
		tools: [],
		approvals: [],
		attachments: [],
		streaming: false,
		error: undefined
	};
}

export function fromHistory(messages: StoredMessage[]): ChatState {
	const items = messages
		.filter((m) => m.role === 'user' || m.role === 'assistant')
		.map((m) => {
			const item = blankItem(`m${m.seq}`, m.role as ChatItem['role'], m.content);
			if (m.answer) {
				item.answer = m.answer;
			}
			if (m.thinking) {
				item.thinking = m.thinking;
			}
			if (m.attachments && m.attachments.length > 0) {
				item.attachments = m.attachments;
			}
			if (m.tools && m.tools.length > 0) {
				item.tools = m.tools.map((t) => ({
					callId: t.call_id,
					name: t.name,
					args: t.args,
					status: t.status,
					outputPreview: t.output_preview,
					durationMs: t.duration_ms
				}));
			}
			return item;
		});
	return { ...initialState(), items };
}

export function addUser(
	state: ChatState,
	text: string,
	attachments: Attachment[] = []
): ChatState {
	const id = `local-${state.items.length}`;
	const item = { ...blankItem(id, 'user', text), attachments };
	return { ...state, items: [...state.items, item], error: null };
}

export function beginRun(state: ChatState, runId: string): ChatState {
	return {
		...ensureLive({ ...state, runId, running: true, lastSeq: 0, error: null, stop: null })
	};
}

function ensureLive(state: ChatState): ChatState {
	const last = state.items[state.items.length - 1];
	if (last && last.role === 'assistant' && last.streaming) return state;
	const live = { ...blankItem(`live-${state.items.length}`, 'assistant'), streaming: true };
	return { ...state, items: [...state.items, live] };
}

function patchLive(state: ChatState, patch: (item: ChatItem) => ChatItem): ChatState {
	const ready = ensureLive(state);
	const index = ready.items.length - 1;
	const items = ready.items.slice();
	items[index] = patch(items[index]);
	return { ...ready, items };
}

export function totalTokens(usage: Usage | undefined): number {
	if (!usage) return 0;
	if (typeof usage.total_tokens === 'number') return usage.total_tokens;
	const keys = ['input_tokens', 'output_tokens', 'prompt_tokens', 'completion_tokens'];
	return keys.reduce((sum, key) => sum + (typeof usage[key] === 'number' ? usage[key] : 0), 0);
}

function settle(state: ChatState): ChatState {
	const items = state.items.map((item) => (item.streaming ? { ...item, streaming: false } : item));
	return { ...state, items, running: false };
}

export function failRun(state: ChatState, message: string): ChatState {
	const settled = settle(state);
	const items = settled.items.slice();
	for (let i = items.length - 1; i >= 0; i--) {
		if (items[i].role === 'assistant') {
			items[i] = { ...items[i], error: message };
			break;
		}
	}
	return { ...settled, items, error: message };
}

export function setApproval(state: ChatState, callId: string, next: ApprovalState): ChatState {
	const items = state.items.map((item) =>
		item.approvals.some((a) => a.callId === callId)
			? {
					...item,
					approvals: item.approvals.map((a) => (a.callId === callId ? { ...a, state: next } : a))
				}
			: item
	);
	return { ...state, items };
}

export function applyEvent(state: ChatState, event: SeqEvent): ChatState {
	if (event.seq > 0 && event.seq <= state.lastSeq) return state;
	const seen = event.seq > 0 ? { ...state, lastSeq: event.seq } : state;

	switch (event.type) {
		case 'RUN_STARTED':
			return ensureLive({ ...seen, runId: event.data.run_id, running: true });
		case 'TEXT_MESSAGE_CONTENT':
			return patchLive(seen, (item) => ({
				...item,
				text: item.text + event.data.delta,
				answer: item.answer + event.data.delta
			}));
		case 'THINKING':
			return patchLive(seen, (item) => ({ ...item, thinking: item.thinking + event.data.delta }));
		case 'TOOL_CALL_START':
			return patchLive(seen, (item) => ({
				...item,
				answer: '',
				tools: [
					...item.tools.filter((t) => t.callId !== event.data.call_id),
					{
						callId: event.data.call_id,
						name: event.data.name,
						args: event.data.args,
						status: 'running'
					}
				]
			}));
		case 'TOOL_CALL_END':
			return patchLive(seen, (item) => ({
				...item,
				tools: item.tools.map((t) =>
					t.callId === event.data.call_id
						? {
								...t,
								status: event.data.ok ? 'ok' : 'failed',
								outputPreview: event.data.output_preview,
								durationMs: event.data.duration_ms
							}
						: t
				),
				approvals: item.approvals.map((a) =>
					a.callId === event.data.call_id && (a.state === 'pending' || a.state === 'sending')
						? { ...a, state: event.data.ok ? 'approved' : 'denied' }
						: a
				)
			}));
		case 'INTERRUPT':
			return patchLive(seen, (item) => ({
				...item,
				approvals: [
					...item.approvals.filter((a) => a.callId !== event.data.call_id),
					{
						callId: event.data.call_id,
						tool: event.data.tool,
						args: event.data.args,
						state: 'pending',
						grantable: event.data.grantable
					}
				]
			}));
		case 'RUN_FINISHED':
			return {
				...settle(seen),
				tokens: seen.tokens + totalTokens(event.data.usage),
				costUsd: seen.costUsd + (event.data.cost_usd ?? 0),
				stop: event.data.stop ?? null
			};
		case 'RUN_ERROR':
			return failRun(seen, event.data.message);
	}
}

export function isTerminal(event: SeqEvent): boolean {
	return event.type === 'RUN_FINISHED' || event.type === 'RUN_ERROR';
}
