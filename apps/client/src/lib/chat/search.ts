import { apiFetch } from '$lib/backend';
import { withSpace } from '$lib/spaces.svelte';

export type SessionHit = { session_id: string; title: string; snippet: string; updated_at: string };

export const SEARCH_DELAY_MS = 250;

/** Says whether a query is long enough to send: two characters once trimmed. */
export function shouldSearch(query: string): boolean {
	return query.trim().length >= 2;
}

export async function searchSessions(query: string, spaceId: string | null | undefined): Promise<SessionHit[]> {
	const path = withSpace(`/sessions/search?q=${encodeURIComponent(query.trim())}`, spaceId);
	return (await apiFetch<{ results: SessionHit[] | null }>(path)).results ?? [];
}

export type Debounced<A extends unknown[]> = { call: (...args: A) => void; cancel: () => void };

/** Delays `fn` until calls stop for `ms`; only the last call's arguments run. */
export function debounce<A extends unknown[]>(fn: (...args: A) => void, ms: number): Debounced<A> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	return {
		call(...args: A) {
			clearTimeout(timer);
			timer = setTimeout(() => fn(...args), ms);
		},
		cancel() {
			clearTimeout(timer);
			timer = undefined;
		}
	};
}
