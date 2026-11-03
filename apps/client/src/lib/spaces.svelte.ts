import { apiFetch } from '$lib/backend';

export type Space = {
	id: string;
	name: string;
	icon: string;
	mycelium_space_id: string;
	opus_workspace: string;
	agenda_calendar: number;
	courrier_account: number;
	registre_space_id: string;
	created_at: string;
	updated_at: string;
};

export type SpaceInput = Partial<
	Pick<Space, 'name' | 'icon' | 'mycelium_space_id' | 'opus_workspace' | 'agenda_calendar' | 'courrier_account'>
>;

const STORAGE_KEY = 'bulle.space';

export function listSpaces(): Promise<Space[]> {
	return apiFetch<Space[]>('/spaces');
}

export function createSpace(input: SpaceInput): Promise<Space> {
	return apiFetch<Space>('/spaces', { method: 'POST', body: JSON.stringify(input) });
}

export function updateSpace(id: string, input: SpaceInput): Promise<Space> {
	return apiFetch<Space>(`/spaces/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify(input)
	});
}

export function deleteSpace(id: string): Promise<void> {
	return apiFetch<void>(`/spaces/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/** Keeps the remembered space when it still exists, otherwise falls back to the default (oldest) one. */
export function pickSpace(items: readonly Space[], remembered: string | null): string | null {
	if (remembered && items.some((s) => s.id === remembered)) return remembered;
	return items[0]?.id ?? null;
}

/** Appends `space_id` to a path when a space is chosen. */
export function withSpace(path: string, spaceId: string | null | undefined): string {
	if (!spaceId) return path;
	const sep = path.includes('?') ? '&' : '?';
	return `${path}${sep}space_id=${encodeURIComponent(spaceId)}`;
}

function remembered(): string | null {
	try {
		return localStorage.getItem(STORAGE_KEY);
	} catch {
		return null;
	}
}

function remember(id: string | null): void {
	try {
		if (id) localStorage.setItem(STORAGE_KEY, id);
		else localStorage.removeItem(STORAGE_KEY);
	} catch {
		return;
	}
}

class SpacesStore {
	items = $state.raw<Space[]>([]);
	currentId = $state<string | null>(null);
	loaded = $state(false);
	current = $derived(this.items.find((s) => s.id === this.currentId) ?? null);

	async load(): Promise<void> {
		try {
			this.set(await listSpaces());
		} catch {
			this.items = [];
		} finally {
			this.loaded = true;
		}
	}

	select(id: string | null): void {
		this.currentId = id;
		remember(id);
	}

	set(items: Space[]): void {
		this.items = items;
		this.select(pickSpace(items, this.currentId ?? remembered()));
	}

	async create(input: SpaceInput): Promise<Space> {
		const created = await createSpace(input);
		this.items = [...this.items, created];
		this.select(created.id);
		return created;
	}

	async update(id: string, input: SpaceInput): Promise<Space> {
		const updated = await updateSpace(id, input);
		this.items = this.items.map((s) => (s.id === id ? updated : s));
		return updated;
	}

	async remove(id: string): Promise<void> {
		await deleteSpace(id);
		this.set(this.items.filter((s) => s.id !== id));
	}
}

export const spaces = new SpacesStore();
