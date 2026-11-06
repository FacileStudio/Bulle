import { countNew, getIdeas, type Goal, type Idea, type IdeaSettings } from './ideas';
import { spaces } from './spaces.svelte';

const POLL_MS = 5000;

export const ideasActivity = $state({
	ideas: [] as Idea[],
	goals: [] as Goal[],
	settings: null as IdeaSettings | null,
	generating: false,
	lastBatchAt: null as string | null,
	spaceId: null as string | null,
	error: null as string | null,
	loaded: false
});

/** Counts the current space's fresh ideas, for the sidebar dot. */
export function newIdeasCount(): number {
	return countNew(ideasActivity.ideas, Date.now());
}

/**
 * Reloads the current space's ideas, goals and settings. A failed request keeps the last good
 * state, unless the space changed, in which case the old space's ideas are cleared.
 */
export async function refreshIdeas(): Promise<void> {
	const spaceId = spaces.currentId;
	if (ideasActivity.spaceId !== spaceId) {
		ideasActivity.ideas = [];
		ideasActivity.goals = [];
		ideasActivity.settings = null;
		ideasActivity.generating = false;
		ideasActivity.lastBatchAt = null;
		ideasActivity.error = null;
		ideasActivity.loaded = false;
		ideasActivity.spaceId = spaceId;
	}
	try {
		const overview = await getIdeas(spaceId);
		if (spaces.currentId !== spaceId) return;
		ideasActivity.ideas = overview.ideas;
		ideasActivity.goals = overview.goals;
		ideasActivity.settings = overview.settings;
		ideasActivity.generating = overview.generating;
		ideasActivity.lastBatchAt = overview.last_batch_at;
		ideasActivity.error = null;
	} catch (err) {
		if (spaces.currentId === spaceId)
			ideasActivity.error = err instanceof Error ? err.message : 'Could not load ideas';
	} finally {
		if (spaces.currentId === spaceId) ideasActivity.loaded = true;
	}
}

/** Polls every 5 s while a batch is generating, and stops once it is done. Returns the cleanup. */
export function pollWhileGenerating(): () => void {
	let timer: ReturnType<typeof setTimeout> | undefined;
	let stopped = false;
	const tick = async () => {
		await refreshIdeas();
		if (!stopped && ideasActivity.generating) timer = setTimeout(tick, POLL_MS);
	};
	timer = setTimeout(tick, POLL_MS);
	return () => {
		stopped = true;
		clearTimeout(timer);
	};
}
