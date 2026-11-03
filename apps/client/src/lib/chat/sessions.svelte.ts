import { api, type SessionStatus, type SessionSummary } from '$lib/api';
import { applyLiveEvent, type LiveEvent } from '$lib/live/events';
import { spaces } from '$lib/spaces.svelte';

class SessionsStore {
	items = $state.raw<SessionSummary[]>([]);
	loaded = $state(false);
	error = $state<string | null>(null);

	async refresh(): Promise<void> {
		try {
			const list = await api().listSessions(spaces.currentId);
			this.items = [...list].sort((a, b) => Date.parse(b.updated_at) - Date.parse(a.updated_at));
			this.error = null;
		} catch (err) {
			this.error = err instanceof Error ? err.message : 'Could not load sessions';
		} finally {
			this.loaded = true;
		}
	}

	/** Refetches the list after the live stream reconnects, once spaces are known. */
	async resync(): Promise<void> {
		if (spaces.loaded) await this.refresh();
	}

	/** Applies a change another device or the server made. */
	apply(event: LiveEvent): void {
		this.items = applyLiveEvent(this.items, event, spaces.currentId);
	}

	/** The status of one session, idle when it is not in the list. */
	statusOf(id: string | null | undefined): SessionStatus {
		return this.items.find((s) => s.id === id)?.status ?? 'idle';
	}

	async create(title?: string): Promise<string> {
		const created = await api().createSession(title, spaces.currentId);
		await this.refresh();
		return created.id;
	}

	async remove(id: string): Promise<void> {
		await api().deleteSession(id);
		this.items = this.items.filter((s) => s.id !== id);
	}
}

export const sessions = new SessionsStore();
