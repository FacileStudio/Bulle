import { t } from '$lib/i18n/index.svelte';
import type { PlanLimitError } from '@facile/muse';
import { ApiError } from '$lib/backend';
import { api, type Decision, type PageContext, type SeqEvent, type SessionDetail, type StreamHandle } from '$lib/api';
import type { LiveEvent } from '$lib/live/events';
import type { Attachment } from './attachments';
import { computer } from './computer.svelte';
import {
	addUser,
	applyEvent,
	beginRun,
	fromHistory,
	initialState,
	isTerminal,
	setApproval,
	type ChatState
} from './reducer';
import { sessions } from './sessions.svelte';

export type Connection = 'idle' | 'live' | 'reconnecting' | 'lost';

function message(err: unknown, fallback: string): string {
	return err instanceof Error && err.message ? err.message : fallback;
}

export class ChatController {
	state = $state.raw<ChatState>(initialState());
	title = $state('');
	channel = $state('');
	loading = $state(false);
	loadError = $state<string | null>(null);
	planLimit = $state<PlanLimitError | null>(null);
	connection = $state<Connection>('idle');
	private sessionId: string | null = null;
	private handle: StreamHandle | null = null;
	private generation = 0;
	private sending = false;
	private queue: SeqEvent[] = [];
	private frame = 0;

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
		this.planLimit = null;
		this.loading = true;
		try {
			const detail = await api().getSession(id);
			if (generation !== this.generation) return;
			this.loading = false;
			this.show(detail);
		} catch (err) {
			if (generation !== this.generation) return;
			this.loadError = message(err, t('chat.loadFailed'));
			this.loading = false;
		}
	}

	private show(detail: SessionDetail): void {
		this.drop();
		this.state = fromHistory(detail.messages);
		this.title = detail.session.title;
		this.channel = detail.session.channel ?? '';
		if (detail.session.unread && this.sessionId) void sessions.markRead(this.sessionId);
		if (detail.active_run_id) this.attach(detail.active_run_id);
		else this.release();
	}

	private async reload(): Promise<void> {
		const id = this.sessionId;
		if (!id || this.loading || this.loadError) return;
		const generation = this.generation;
		try {
			const detail = await api().getSession(id);
			if (generation === this.generation && id === this.sessionId) this.show(detail);
		} catch {
			return;
		}
	}

	/** Reloads the open conversation after the live stream reconnects. */
	resync(): void {
		void this.reload();
	}

	/** Follows what another device or the server did to the open conversation. */
	applyLive(event: LiveEvent): void {
		if (event.type === 'notification' || event.data.session_id !== this.sessionId) return;
		this.settle();
		switch (event.type) {
			case 'session.updated':
				this.title = event.data.title;
				return;
			case 'session.deleted':
				this.close();
				this.loadError = t('chat.deleted');
				return;
			case 'run.started':
				if (!this.sending && this.state.runId !== event.data.run_id) void this.reload();
				return;
			case 'run.finished':
				void this.reload();
				return;
			case 'approval.decided':
				if (this.state.runId === event.data.run_id && event.data.approved !== undefined) {
					this.state = setApproval(this.state, event.data.call_id, event.data.approved ? 'approved' : 'denied');
				}
		}
	}

	close(): void {
		this.drop();
		this.generation++;
		this.handle?.close();
		this.handle = null;
		this.connection = 'idle';
	}

	private attach(runId: string): void {
		this.drop();
		this.handle?.close();
		this.state = beginRun(this.state, runId);
		this.connection = 'reconnecting';
		this.handle = api().openRunStream(runId, 0, {
			onOpen: () => {
				const wasLost = this.connection === 'lost';
				this.connection = 'live';
				if (wasLost) void this.reload();
			},
			onReconnecting: () => {
				if (this.connection !== 'lost') this.connection = 'reconnecting';
			},
			onLost: () => {
				this.connection = 'lost';
			},
			onEvent: (event) => this.enqueue(event)
		});
	}

	/** Queues a streamed event and folds the queue into state once per animation frame. */
	private enqueue(event: SeqEvent): void {
		this.queue.push(event);
		if (isTerminal(event) || typeof requestAnimationFrame !== 'function') {
			this.settle();
			return;
		}
		if (!this.frame) {
			this.frame = requestAnimationFrame(() => {
				this.frame = 0;
				this.settle();
			});
		}
	}

	/** Applies every queued event now, so a read or a write sees the latest state. */
	private settle(): void {
		this.cancelFrame();
		if (!this.queue.length) return;
		const events = this.queue;
		this.queue = [];
		let next = this.state;
		for (const event of events) next = applyEvent(next, event);
		this.state = next;
		if (events.some(isTerminal)) {
			this.release();
			void sessions.refresh();
		}
	}

	/** Forgets queued events that belong to a stream this controller left. */
	private drop(): void {
		this.cancelFrame();
		this.queue = [];
	}

	private cancelFrame(): void {
		if (!this.frame) return;
		cancelAnimationFrame(this.frame);
		this.frame = 0;
	}

	private release(): void {
		this.handle?.close();
		this.handle = null;
		this.connection = 'idle';
	}

	async send(text: string, attachments: Attachment[] = [], page: PageContext | null = null): Promise<boolean> {
		const id = this.sessionId;
		const trimmed = text.trim();
		this.settle();
		if (!id || (!trimmed && !attachments.length) || this.state.running) return false;
		this.sending = true;
		this.planLimit = null;
		try {
			const ids = attachments.map((a) => a.id);
			const { run_id } = await api().sendMessage(id, trimmed, ids, page);
			if (id !== this.sessionId) return true;
			this.state = addUser(this.state, trimmed, attachments);
			this.attach(run_id);
			return true;
		} catch (err) {
			if (err instanceof ApiError && err.planLimit) {
				this.planLimit = err.planLimit;
				return false;
			}
			this.state = { ...this.state, error: message(err, t('chat.sendFailed')) };
			return false;
		} finally {
			this.sending = false;
		}
	}

	async stop(): Promise<void> {
		this.settle();
		const runId = this.state.runId;
		if (!runId || !this.state.running) return;
		try {
			await api().stopRun(runId);
		} catch (err) {
			this.state = { ...this.state, error: message(err, t('chat.stopFailed')) };
		}
	}

	async decide(callId: string, decision: Decision): Promise<void> {
		this.settle();
		const runId = this.state.runId;
		if (!runId) return;
		this.state = setApproval(this.state, callId, 'sending');
		try {
			await api().decide(runId, callId, decision);
			this.state = setApproval(this.state, callId, decision.approve ? 'approved' : 'denied');
		} catch (err) {
			this.state = {
				...this.rollback(callId),
				error: message(err, t('chat.decideFailed'))
			};
		}
	}

	private callState(callId: string): string | undefined {
		for (const item of this.state.items) {
			const found = item.approvals.find((a) => a.callId === callId);
			if (found) return found.state;
		}
		return undefined;
	}

	private rollback(callId: string): ChatState {
		return this.callState(callId) === 'sending'
			? setApproval(this.state, callId, 'pending')
			: this.state;
	}

	async decideAll(requested: string[], decision: Decision): Promise<void> {
		this.settle();
		const runId = this.state.runId;
		const callIds = requested.filter((id) => this.callState(id) === 'pending');
		if (!runId || callIds.length === 0) return;
		for (const callId of callIds) this.state = setApproval(this.state, callId, 'sending');
		const results = await Promise.allSettled(callIds.map((callId) => api().decide(runId, callId, decision)));
		let failure: unknown = null;
		results.forEach((result, index) => {
			if (result.status === 'fulfilled') {
				this.state = setApproval(this.state, callIds[index], decision.approve ? 'approved' : 'denied');
				return;
			}
			failure = result.reason;
			this.state = this.rollback(callIds[index]);
		});
		if (failure) {
			this.state = { ...this.state, error: message(failure, t('chat.decideFailed')) };
		}
	}

	dismissError(): void {
		this.settle();
		this.planLimit = null;
		this.state = { ...this.state, error: null };
	}
}

export const chat = new ChatController();
