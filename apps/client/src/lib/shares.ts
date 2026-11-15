import { ApiError, apiFetch } from '$lib/backend';
import { artifactPath, type ArtifactWhere } from '$lib/artifacts';

export type Share = { token: string; url: string; created_at: string };

export type PublicShare = { title: string; content: string; updated: string };

/** The API path of an artifact's public link, keeping the query that says where it lives. */
export function sharePath(id: string, where: ArtifactWhere): string {
	const [path, query] = artifactPath(id, where).split('?');
	return query ? `${path}/share?${query}` : `${path}/share`;
}

/** The live public link, or null when the artifact has none. */
export async function getShare(id: string, where: ArtifactWhere): Promise<Share | null> {
	try {
		return await apiFetch<Share>(sharePath(id, where));
	} catch (err) {
		if (err instanceof ApiError && err.status === 404) return null;
		throw err;
	}
}

/** Makes the public link, or answers the one that is already live. */
export function createShare(id: string, where: ArtifactWhere): Promise<Share> {
	return apiFetch<Share>(sharePath(id, where), { method: 'POST' });
}

export function revokeShare(id: string, where: ArtifactWhere): Promise<void> {
	return apiFetch<void>(sharePath(id, where), { method: 'DELETE' });
}

/** Reads what a public link shows: no session, read only. */
export function getPublicShare(token: string): Promise<PublicShare> {
	return apiFetch<PublicShare>(`/public/shares/${encodeURIComponent(token)}`);
}
