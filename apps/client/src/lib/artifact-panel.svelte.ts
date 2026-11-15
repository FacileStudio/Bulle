import type { ArtifactRef } from './artifacts';

export const artifactPanel = $state<{ ref: ArtifactRef | null; auto: boolean }>({ ref: null, auto: false });

const closed = new Set<string>();

/** The ids the reader closed since the last run began; a later write to one does not reopen it. */
export function closedArtifacts(): ReadonlySet<string> {
	return closed;
}

/** Forgets which artifacts the reader closed, when a new run begins. */
export function resetClosed(): void {
	closed.clear();
}

/** Opens the panel on an artifact the assistant just wrote, without moving focus. */
export function autoOpenArtifact(ref: ArtifactRef): void {
	opener = null;
	artifactPanel.auto = true;
	artifactPanel.ref = ref;
}

let opener: HTMLElement | null = null;

/** Opens the side panel on one artifact and remembers what to give focus back to. */
export function openArtifact(ref: ArtifactRef, from?: HTMLElement | null): void {
	opener = from ?? null;
	artifactPanel.auto = false;
	artifactPanel.ref = ref;
}

/** Closes the panel and returns focus to the card that opened it, when it is still there. */
export function closeArtifact(): void {
	if (artifactPanel.ref) closed.add(artifactPanel.ref.id);
	artifactPanel.ref = null;
	const target = opener;
	opener = null;
	if (target?.isConnected) queueMicrotask(() => target.focus());
}
