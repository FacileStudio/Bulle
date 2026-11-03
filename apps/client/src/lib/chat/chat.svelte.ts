import { api, type StreamHandle } from '$lib/api';
import type { Attachment } from './attachments';
import { computer } from './computer.svelte';
import {
	addUser,
	applyEvent,
	beginRun,
	failRun,
	fromHistory,
	initialState,
	isTerminal,
	setApproval,
	type ChatState
} from './reducer';
import { sessions } from './sessions.svelte';

export type Connection = 'idle' | 'live' | 'reconnecting';

function message(err: unknown, fallback: string): string {
	return err instanceof Error && err.message ? err.message : fallback;
}

export class ChatController {
	state = $state.raw<ChatState>(initialState());
	title = $state('');
	channel = $state('');
	loading = $state(false);
	loadError = $state<string | null>(null);
	connection = $state<Connection>('idle');
	private sessionId: string | null = null;
	private handle: StreamHandle | null = null;
	private generation = 0;

	get current(): string | null {
		return this.sessionId;
	}

	async open(id: string): Promise<void> {
		this.close();
		const generation = ++this.generation;
		this.sessionId = id;
		this.state = initialState();
		this.title = '';
		this.channel = '';
		this.loadError = null;
		this.loading = true;
		try {
			const detail = await api().getSession(id);
			if (generation !== this.generation) return;
			this.state = fromHistory(detail.messages);
			this.title = detail.session.title;
			this.channel = detail.session.channel ?? '';
			this.loading = false;
			if (detail.active_run_id) this.attach(detail.active_run_id);
		} catch (err) {
			if (generation !== this.generation) return;
			this.loadError = message(err, 'Could not load this session');
			this.loading = false;
		}
	}

	close(): void {
		this.generation++;
		this.handle?.close();
		this.handle = null;
		this.connection = 'idle';
	}

	private attach(runId: string): void {
		this.handle?.close();
		this.state = beginRun(this.state, runId);
		this.connection = 'reconnecting';
		this.handle = api().openRunStream(runId, 0, {
			onOpen: () => {
				this.connection = 'live';
			},
			onReconnecting: () => {
				this.connection = 'reconnecting';
			},
			onFatal: (text) => {
				this.state = failRun(this.state, text);
				this.release();
			},
			onEvent: (event) => {
				this.state = applyEvent(this.state, event);
				if (isTerminal(event)) {
					this.release();
					void sessions.refresh();
				}
			}
		});
	}

	private release(): void {
		this.handle?.close();
		this.handle = null;
		this.connection = 'idle';
	}

	async send(text: string, attachments: Attachment[] = []): Promise<boolean> {
		const id = this.sessionId;
		const trimmed = text.trim();
		if (!id || (!trimmed && !attachments.length) || this.state.running) return false;
		try {
			const ids = attachments.map((a) => a.id);
			const { run_id } = await api().sendMessage(id, trimmed, ids);
			if (id !== this.sessionId) return true;
			this.state = addUser(this.state, trimmed, attachments);
			this.attach(run_id);
			return true;
		} catch (err) {
			this.state = { ...this.state, error: message(err, 'Could not send the message') };
			return false;
		}
	}

	async stop(): Promise<void> {
		const runId = this.state.runId;
		if (!runId || !this.state.running) return;
		try {
			await api().stopRun(runId);
		} catch (err) {
			this.state = { ...this.state, error: message(err, 'Could not stop the run') };
		}
	}

	async decide(callId: string, approve: boolean): Promise<void> {
		const runId = this.state.runId;
		if (!runId) return;
		this.state = setApproval(this.state, callId, 'sending');
		try {
			await api().decide(runId, callId, approve);
			this.state = setApproval(this.state, callId, approve ? 'approved' : 'denied');
		} catch (err) {
			this.state = {
				...setApproval(this.state, callId, 'pending'),
				error: message(err, 'Could not send your decision')
			};
		}
	}

	dismissError(): void {
		this.state = { ...this.state, error: null };
	}
}

export const chat = new ChatController();
