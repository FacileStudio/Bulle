import { t } from '$lib/i18n/index.svelte';
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

let latest = 0;

/** Counts the current space's fresh ideas, for the sidebar dot. */
export function newIdeasCount(): number {
	return countNew(ideasActivity.ideas, Date.now());
}

/** Says whether a response for `spaceId` still belongs on screen. */
export function isCurrentSpace(spaceId: string | null): boolean {
	return spaces.currentId === spaceId && ideasActivity.spaceId === spaceId;
}

/**
 * Reloads the current space's ideas, goals and settings. Only the newest request writes its
 * answer, so a slow reply for an old space or an older poll never overwrites a newer one. A
 * failed request keeps the last good state, unless the space changed, which clears it first.
 */
export async function refreshIdeas(): Promise<void> {
	const spaceId = spaces.currentId;
	const request = ++latest;
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
		if (request !== latest || !isCurrentSpace(spaceId)) return;
		ideasActivity.ideas = overview.ideas;
		ideasActivity.goals = overview.goals;
		ideasActivity.settings = overview.settings;
		ideasActivity.generating = overview.generating;
		ideasActivity.lastBatchAt = overview.last_batch_at;
		ideasActivity.error = null;
		ideasActivity.loaded = true;
	} catch (err) {
		if (request !== latest || !isCurrentSpace(spaceId)) return;
		ideasActivity.error = err instanceof Error ? err.message : t('ideas.loadError');
		ideasActivity.loaded = true;
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
