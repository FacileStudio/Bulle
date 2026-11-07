import { ApiError, apiFetch } from '$lib/backend';
import { errorText } from '$lib/ideas';
import { t } from '$lib/i18n/index.svelte';
import { withSpace } from '$lib/spaces.svelte';

export type MemoryPage = { path: string; size: number; updated: string };

export type MemoryList = { pages: MemoryPage[]; available: boolean };

export type MemoryPageDetail = { path: string; content: string };

export const PROFILE_PAGE = 'profile.md';

/** Builds the API path of one memory page, escaping each segment so a slash stays a folder. */
export function memoryPagePath(path: string, spaceId: string | null | undefined): string {
	const escaped = path.split('/').map(encodeURIComponent).join('/');
	return withSpace(`/memory/pages/${escaped}`, spaceId);
}

export async function listMemoryPages(spaceId: string | null | undefined): Promise<MemoryList> {
	const list = await apiFetch<Partial<MemoryList>>(withSpace('/memory/pages', spaceId));
	return { pages: list.pages ?? [], available: list.available === true };
}

export function getMemoryPage(path: string, spaceId: string | null | undefined): Promise<MemoryPageDetail> {
	return apiFetch<MemoryPageDetail>(memoryPagePath(path, spaceId));
}

export function saveMemoryPage(
	path: string,
	content: string,
	spaceId: string | null | undefined
): Promise<MemoryPageDetail> {
	return apiFetch<MemoryPageDetail>(memoryPagePath(path, spaceId), {
		method: 'PUT',
		body: JSON.stringify({ content })
	});
}

export function deleteMemoryPage(path: string, spaceId: string | null | undefined): Promise<void> {
	return apiFetch<void>(memoryPagePath(path, spaceId), { method: 'DELETE' });
}

/** Names a page for the list: the profile by its role, any other page by its path without .md. */
export function pageLabel(path: string): string {
	return path === PROFILE_PAGE ? t('memory.profile') : path.replace(/\.md$/, '');
}

/** Turns a memory request failure into a sentence, and says when Mycelium needs connecting. */
export function memoryError(err: unknown, fallback: string): { text: string; connect: boolean } {
	const status = err instanceof ApiError ? err.status : 0;
	if (status === 412) return { text: t('memory.error.connect'), connect: true };
	if (status === 404) return { text: t('memory.error.missing'), connect: false };
	if (status === 503) return { text: t('artifacts.error.unavailable'), connect: false };
	return { text: errorText(err, fallback), connect: false };
}
