import type { ArtifactRef } from './artifacts';

export const artifactPanel = $state<{ ref: ArtifactRef | null }>({ ref: null });

let opener: HTMLElement | null = null;

/** Opens the side panel on one artifact and remembers what to give focus back to. */
export function openArtifact(ref: ArtifactRef, from?: HTMLElement | null): void {
	opener = from ?? null;
	artifactPanel.ref = ref;
}

/** Closes the panel and returns focus to the card that opened it, when it is still there. */
export function closeArtifact(): void {
	artifactPanel.ref = null;
	const target = opener;
	opener = null;
	if (target?.isConnected) queueMicrotask(() => target.focus());
}
