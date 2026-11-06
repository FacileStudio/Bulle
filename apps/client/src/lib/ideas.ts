import { apiFetch } from '$lib/backend';
import { withSpace } from '$lib/spaces.svelte';

export type IdeaKind = 'mail' | 'task' | 'event' | 'follow_up' | 'research' | 'other';
export type IdeaStatus = 'new' | 'saved' | 'done' | 'dismissed' | 'expired';
export type IdeaFeedback = '' | 'up' | 'down';

export type Idea = {
	id: string;
	space_id: string;
	batch_id: string;
	kind: IdeaKind;
	title: string;
	why: string;
	prompt: string;
	draft: string;
	sources: { label: string; ref: string }[];
	status: IdeaStatus;
	feedback: IdeaFeedback;
	session_id: string;
	created_at: string;
	expires_at: string;
};

export const GOAL_CATEGORIES = [
	'health',
	'relationships',
	'finance',
	'career',
	'interests',
	'productivity',
	'other'
] as const;

export type GoalCategory = (typeof GOAL_CATEGORIES)[number];

export type Goal = {
	id: string;
	space_id: string;
	title: string;
	category: GoalCategory;
	done: boolean;
	created_at: string;
};

export type IdeaSettings = {
	space_id: string;
	enabled: boolean;
	hour: number;
	timezone: string;
	push: boolean;
	steering: string;
};

export type IdeasOverview = {
	ideas: Idea[];
	goals: Goal[];
	settings: IdeaSettings;
	generating: boolean;
	last_batch_at: string | null;
};

export type IdeaPatch = {
	status?: 'saved' | 'new' | 'dismissed';
	feedback?: IdeaFeedback;
	reason?: string;
};

export type SettingsPatch = Partial<Omit<IdeaSettings, 'space_id'>>;

export type GoalPatch = { title?: string; category?: GoalCategory; done?: boolean };

export const DISMISS_REASONS = ['Not relevant', 'Already done', 'Wrong time'] as const;

function send(method: string, body: unknown): RequestInit {
	return { method, body: JSON.stringify(body) };
}

function ideaPath(id: string, suffix = ''): string {
	return `/ideas/${encodeURIComponent(id)}${suffix}`;
}

function goalPath(id: string): string {
	return `/ideas/goals/${encodeURIComponent(id)}`;
}

export function getIdeas(spaceId: string | null | undefined): Promise<IdeasOverview> {
	return apiFetch<IdeasOverview>(withSpace('/ideas', spaceId));
}

export function updateIdea(id: string, patch: IdeaPatch): Promise<Idea> {
	return apiFetch<Idea>(ideaPath(id), send('PATCH', patch));
}

export function startIdea(id: string): Promise<{ session_id: string }> {
	return apiFetch<{ session_id: string }>(ideaPath(id, '/start'), { method: 'POST' });
}

export function generateIdeas(spaceId: string): Promise<{ batch_id: string }> {
	return apiFetch<{ batch_id: string }>('/ideas/generate', send('POST', { space_id: spaceId }));
}

export function updateIdeaSettings(spaceId: string, patch: SettingsPatch): Promise<IdeaSettings> {
	return apiFetch<IdeaSettings>('/ideas/settings', send('PUT', { space_id: spaceId, ...patch }));
}

export function createGoal(spaceId: string, title: string, category: GoalCategory): Promise<Goal> {
	return apiFetch<Goal>('/ideas/goals', send('POST', { space_id: spaceId, title, category }));
}

export function updateGoal(id: string, patch: GoalPatch): Promise<Goal> {
	return apiFetch<Goal>(goalPath(id), send('PUT', patch));
}

export function deleteGoal(id: string): Promise<void> {
	return apiFetch<void>(goalPath(id), { method: 'DELETE' });
}

/** Says whether an idea is still on offer at `now`: new or saved, and not past its expiry. */
export function isActive(idea: Idea, now: number): boolean {
	if (idea.status !== 'new' && idea.status !== 'saved') return false;
	const expires = Date.parse(idea.expires_at);
	return idea.status === 'saved' || !Number.isFinite(expires) || expires > now;
}

/** Counts the fresh ideas the sidebar dot announces. */
export function countNew(ideas: readonly Idea[], now: number): number {
	return ideas.filter((idea) => idea.status === 'new' && isActive(idea, now)).length;
}

/** Splits active ideas into fresh ones, newest batch first, and saved ones. */
export function groupIdeas(ideas: readonly Idea[], now: number): { fresh: Idea[]; saved: Idea[] } {
	const active = ideas.filter((idea) => isActive(idea, now));
	const newest = (a: Idea, b: Idea) => b.created_at.localeCompare(a.created_at);
	return {
		fresh: active.filter((idea) => idea.status === 'new').sort(newest),
		saved: active.filter((idea) => idea.status === 'saved').sort(newest)
	};
}

/** Orders goals open first, then done, oldest first within each. */
export function sortGoals(goals: readonly Goal[]): Goal[] {
	return [...goals].sort(
		(a, b) => Number(a.done) - Number(b.done) || a.created_at.localeCompare(b.created_at)
	);
}

/** Returns the browser time zone when the stored one is still the UTC default and differs. */
export function timezoneToSync(stored: string, browser: string | undefined): string | null {
	return stored === 'UTC' && browser && browser !== 'UTC' ? browser : null;
}

/** Replaces one item by id, keeping the list's order. */
export function replaceById<T extends { id: string }>(items: readonly T[], next: T): T[] {
	return items.map((item) => (item.id === next.id ? next : item));
}
