import { api, type SessionSummary } from '$lib/api';
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
