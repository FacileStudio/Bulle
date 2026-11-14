import { ApiError, apiFetch } from '$lib/backend';
import { errorText } from '$lib/ideas';
import { t } from '$lib/i18n/index.svelte';
import type { ToolCall } from '$lib/chat/reducer';
import { withSpace } from '$lib/spaces.svelte';

export type Artifact = {
	id: string;
	title: string;
	machine: string;
	created: string;
	updated: string;
	mycelium_space?: string;
	space_name?: string;
	session_id?: string;
};

export type ArtifactSpace = { key: string; name: string };

/** Where an artifact lives: the Mycelium tree when known, else the Bulle space that wrote it. */
export type ArtifactWhere = { myceliumSpace?: string | null; spaceId?: string | null };

export const COMMON_TREE = 'common';

export type ArtifactDetail = Artifact & { content: string };

export type ArtifactList = { artifacts: Artifact[]; available: boolean };

export type ArtifactRef = { id: string; title: string; space_id: string; mycelium_space?: string };

export type ArtifactHit = Artifact & { match: 'title' | 'content'; snippet: string };

export type ArtifactSearch = { artifacts: ArtifactHit[]; next_cursor: string; available: boolean };

/** The longest query `GET /api/artifacts/search` accepts, in characters. */
export const MAX_QUERY = 200;

const WRITE_TOOLS = new Set(['artifact_create', 'artifact_update']);

/** The API path of one artifact, carrying where it lives when that is known. */
export function artifactPath(id: string, where: ArtifactWhere): string {
	const path = withSpace(`/artifacts/${encodeURIComponent(id)}`, where.spaceId);
	if (!where.myceliumSpace) return path;
	const sep = path.includes('?') ? '&' : '?';
	return `${path}${sep}mycelium_space=${encodeURIComponent(where.myceliumSpace)}`;
}

/** The page of one artifact; `in` keeps the Mycelium tree so the page opens it directly. */
export function artifactHref(artifact: Pick<Artifact, 'id' | 'mycelium_space'>): string {
	const href = `/artifacts/${encodeURIComponent(artifact.id)}`;
	return artifact.mycelium_space ? `${href}?in=${encodeURIComponent(artifact.mycelium_space)}` : href;
}

export async function listArtifacts(): Promise<ArtifactList> {
	const list = await apiFetch<Partial<ArtifactList>>('/artifacts');
	return { artifacts: list.artifacts ?? [], available: list.available === true };
}

export function getArtifact(id: string, where: ArtifactWhere): Promise<ArtifactDetail> {
	return apiFetch<ArtifactDetail>(artifactPath(id, where));
}

export function deleteArtifact(id: string, where: ArtifactWhere): Promise<void> {
	return apiFetch<void>(artifactPath(id, where), { method: 'DELETE' });
}

/** Replaces an artifact's markdown, and its title when one is given; the API answers the saved artifact. */
export function updateArtifact(id: string, where: ArtifactWhere, body: { title?: string; content: string }): Promise<ArtifactDetail> {
	return apiFetch<ArtifactDetail>(artifactPath(id, where), { method: 'PUT', body: JSON.stringify(body) });
}

/** The name an artifact's space shows under: the space's own, or Personal for the common tree. */
export function artifactSpaceName(artifact: Pick<Artifact, 'mycelium_space' | 'space_name'>): string {
	if (!artifact.mycelium_space || artifact.mycelium_space === COMMON_TREE) return t('artifacts.personal');
	return artifact.space_name || t('artifacts.personal');
}

/** The distinct spaces the artifacts live in, personal first, then by name. */
export function artifactSpaces(items: readonly Artifact[]): ArtifactSpace[] {
	const spaces = new Map<string, ArtifactSpace>();
	for (const item of items) {
		const key = item.mycelium_space || COMMON_TREE;
		if (!spaces.has(key)) spaces.set(key, { key, name: artifactSpaceName(item) });
	}
	return [...spaces.values()].sort((a, b) =>
		a.key === COMMON_TREE ? -1 : b.key === COMMON_TREE ? 1 : a.name.localeCompare(b.name)
	);
}

/** Keeps the artifacts of one space; an empty key keeps them all. */
export function inSpace<T extends Artifact>(items: readonly T[], key: string): T[] {
	if (!key) return [...items];
	return items.filter((a) => (a.mycelium_space || COMMON_TREE) === key);
}

function readRef(output: string | undefined): ArtifactRef | null {
	if (!output) return null;
	try {
		const ref = (JSON.parse(output) as { artifact?: Partial<ArtifactRef> }).artifact;
		if (!ref || typeof ref.id !== 'string' || !ref.id) return null;
		return {
			id: ref.id,
			title: typeof ref.title === 'string' && ref.title.trim() ? ref.title : t('artifacts.untitled'),
			space_id: typeof ref.space_id === 'string' ? ref.space_id : '',
			mycelium_space: typeof ref.mycelium_space === 'string' ? ref.mycelium_space : undefined
		};
	} catch {
		return null;
	}
}

/**
 * Finds the artifacts a reply wrote: one card per finished `artifact_create` or
 * `artifact_update` call, in first-seen order, carrying the latest title for each id.
 */
export function artifactCards(tools: readonly ToolCall[]): ArtifactRef[] {
	const cards = new Map<string, ArtifactRef>();
	for (const call of tools) {
		if (call.status !== 'ok' || !WRITE_TOOLS.has(call.name)) continue;
		const ref = readRef(call.outputPreview);
		if (ref) cards.set(ref.id, ref);
	}
	return [...cards.values()];
}

/** The query the search route receives: trimmed and cut to the length it accepts. */
export function normaliseQuery(raw: string): string {
	return raw.trim().slice(0, MAX_QUERY);
}

/** Searches every artifact the user can read by title and content, one page per cursor. */
export async function searchArtifacts(query: string, cursor = ''): Promise<ArtifactSearch> {
	const params = new URLSearchParams({ q: normaliseQuery(query) });
	if (cursor) params.set('cursor', cursor);
	const page = await apiFetch<Partial<ArtifactSearch>>(`/artifacts/search?${params}`);
	return { artifacts: page.artifacts ?? [], next_cursor: page.next_cursor ?? '', available: page.available === true };
}

export type SnippetPart = { text: string; hit: boolean };

function escapeRegExp(word: string): string {
	return word.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

/** Splits a snippet around each case-insensitive occurrence of a query word, so the page marks what matched without HTML. */
export function snippetParts(text: string, query: string): SnippetPart[] {
	const words = normaliseQuery(query).toLowerCase().split(/\s+/).filter(Boolean);
	if (!text) return [];
	if (words.length === 0) return [{ text, hit: false }];
	const pattern = new RegExp(`(${[...words].sort((a, b) => b.length - a.length).map(escapeRegExp).join('|')})`, 'gi');
	return text
		.split(pattern)
		.filter(Boolean)
		.map((chunk) => ({ text: chunk, hit: words.includes(chunk.toLowerCase()) }));
}

/** True when a key press lands where text is typed, so `/` must type a slash there. */
export function isEditableTarget(target: EventTarget | null): boolean {
	const el = target as HTMLElement | null;
	if (!el || typeof el.tagName !== 'string') return false;
	return el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName);
}

/** Turns an artifact request failure into a sentence for the reader. */
export function artifactError(err: unknown): { text: string; connect: boolean } {
	const status = err instanceof ApiError ? err.status : 0;
	if (status === 412) return { text: t('artifacts.error.connect'), connect: true };
	if (status === 404) return { text: t('artifacts.error.missing'), connect: false };
	if (status === 503) return { text: t('artifacts.error.unavailable'), connect: false };
	return { text: errorText(err, t('artifacts.error.load')), connect: false };
}
