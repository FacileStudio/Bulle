import { t } from '$lib/i18n/index.svelte';
import type { PlanLimitError } from '@facile/muse';
import { ApiError } from '$lib/backend';
import { api, type Decision, type SessionDetail, type StreamHandle } from '$lib/api';
import type { LiveEvent } from '$lib/live/events';
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
	planLimit = $state<PlanLimitError | null>(null);
	connection = $state<Connection>('idle');
	private sessionId: string | null = null;
	private handle: StreamHandle | null = null;
	private generation = 0;
	private sending = false;

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
		this.state = fromHistory(detail.messages);
		this.title = detail.session.title;
		this.channel = detail.session.channel ?? '';
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
		this.sending = true;
		this.planLimit = null;
		try {
			const ids = attachments.map((a) => a.id);
			const { run_id } = await api().sendMessage(id, trimmed, ids);
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
		const runId = this.state.runId;
		if (!runId || !this.state.running) return;
		try {
			await api().stopRun(runId);
		} catch (err) {
			this.state = { ...this.state, error: message(err, t('chat.stopFailed')) };
		}
	}

	async decide(callId: string, decision: Decision): Promise<void> {
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
		this.planLimit = null;
		this.state = { ...this.state, error: null };
	}
}

export const chat = new ChatController();
