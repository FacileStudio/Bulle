import { ApiError, apiFetch } from '$lib/backend';
import { errorText } from '$lib/ideas';
import type { ToolCall } from '$lib/chat/reducer';
import { withSpace } from '$lib/spaces.svelte';

export type Artifact = {
	id: string;
	title: string;
	machine: string;
	created: string;
	updated: string;
	session_id?: string;
};

export type ArtifactDetail = Artifact & { content: string };

export type ArtifactList = { artifacts: Artifact[]; available: boolean };

export type ArtifactRef = { id: string; title: string; space_id: string };

export type ArtifactMonth = { key: string; label: string; items: Artifact[] };

const WRITE_TOOLS = new Set(['artifact_create', 'artifact_update']);

function artifactPath(id: string, spaceId: string | null | undefined): string {
	return withSpace(`/artifacts/${encodeURIComponent(id)}`, spaceId);
}

export async function listArtifacts(spaceId: string | null | undefined): Promise<ArtifactList> {
	const list = await apiFetch<Partial<ArtifactList>>(withSpace('/artifacts', spaceId));
	return { artifacts: list.artifacts ?? [], available: list.available === true };
}

export function getArtifact(id: string, spaceId: string | null | undefined): Promise<ArtifactDetail> {
	return apiFetch<ArtifactDetail>(artifactPath(id, spaceId));
}

export function deleteArtifact(id: string, spaceId: string | null | undefined): Promise<void> {
	return apiFetch<void>(artifactPath(id, spaceId), { method: 'DELETE' });
}

function readRef(output: string | undefined): ArtifactRef | null {
	if (!output) return null;
	try {
		const ref = (JSON.parse(output) as { artifact?: Partial<ArtifactRef> }).artifact;
		if (!ref || typeof ref.id !== 'string' || !ref.id) return null;
		return {
			id: ref.id,
			title: typeof ref.title === 'string' && ref.title.trim() ? ref.title : 'Untitled artifact',
			space_id: typeof ref.space_id === 'string' ? ref.space_id : ''
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

/** Keeps the artifacts whose title holds every word of the query, ignoring case. */
export function filterArtifacts(items: readonly Artifact[], query: string): Artifact[] {
	const words = query.toLowerCase().split(/\s+/).filter(Boolean);
	return items.filter((a) => words.every((w) => a.title.toLowerCase().includes(w)));
}

/** Sorts artifacts newest first and groups them by the month they were last updated. */
export function groupByMonth(items: readonly Artifact[], locale?: string): ArtifactMonth[] {
	const months = new Map<string, ArtifactMonth>();
	const sorted = [...items].sort((a, b) => Date.parse(b.updated) - Date.parse(a.updated));
	for (const item of sorted) {
		const date = new Date(item.updated);
		const valid = !Number.isNaN(date.getTime());
		const key = valid ? `${date.getFullYear()}-${date.getMonth()}` : 'unknown';
		const label = valid
			? date.toLocaleDateString(locale, { month: 'long', year: 'numeric' })
			: 'Undated';
		const month = months.get(key) ?? { key, label, items: [] };
		month.items.push(item);
		months.set(key, month);
	}
	return [...months.values()];
}

/** Turns an artifact request failure into a sentence for the reader. */
export function artifactError(err: unknown): { text: string; connect: boolean } {
	const status = err instanceof ApiError ? err.status : 0;
	if (status === 412) return { text: 'Connect Mycelium in Settings, Apps, to see your artifacts.', connect: true };
	if (status === 404) return { text: 'This artifact does not exist any more.', connect: false };
	if (status === 503) return { text: 'Mycelium is not answering. Try again in a moment.', connect: false };
	return { text: errorText(err, 'Could not load the artifact.'), connect: false };
}
