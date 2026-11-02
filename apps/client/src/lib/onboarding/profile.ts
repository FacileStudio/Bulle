import { ARCHETYPES, RANDOM_NAMES, TRAITS, avatarUrlFor } from '../assistant';

export const DEFAULT_NAME = 'Bulle';
export const DEFAULT_ARCHETYPE = 'pragmatic_senior';
export const DEFAULT_TRAITS: readonly string[] = ['concise'];
export const NAME_MAX_LENGTH = 40;

export type Draft = {
	step: number;
	name: string;
	archetype: string;
	traits: string[];
};

export type ProfileInput = {
	name?: string | null;
	archetype?: string | null;
	traits?: string[] | null;
};

export type ProfilePayload = {
	name: string;
	avatar_url: string;
	archetype: string;
	traits: string[];
};

const TINTS: Record<string, readonly [string, string]> = {
	pragmatic_senior: ['var(--color-fc-chart-3)', 'var(--color-fc-chart-1)'],
	terse_minimalist: ['var(--color-fc-chart-1)', 'var(--color-fc-chart-3)'],
	supportive_mentor: ['var(--color-fc-chart-2)', 'var(--color-fc-chart-5)'],
	witty_hacker: ['var(--color-fc-chart-6)', 'var(--color-fc-chart-1)']
};

export function normalizeArchetype(value: string | null | undefined): string {
	const found = ARCHETYPES.find((a) => a.id === value || a.label === value);
	return found ? found.id : DEFAULT_ARCHETYPE;
}

export function normalizeTrait(value: string): string {
	return TRAITS.find((t) => t.id === value || t.label === value)?.id ?? value;
}

export function hasTrait(traits: readonly string[], id: string): boolean {
	return traits.some((t) => normalizeTrait(t) === id);
}

export function toggleTrait(traits: readonly string[], id: string): string[] {
	if (hasTrait(traits, id)) return traits.filter((t) => normalizeTrait(t) !== id);
	return [...traits.map(normalizeTrait), id];
}

export function randomName(current: string, pick: () => number = Math.random): string {
	const pool = RANDOM_NAMES.filter((n) => n.toLowerCase() !== current.trim().toLowerCase());
	return pool[Math.floor(pick() * pool.length)] ?? RANDOM_NAMES[0];
}

export function initialDraft(profile?: ProfileInput | null): Draft {
	return {
		step: 0,
		name: profile?.name?.trim() ?? '',
		archetype: normalizeArchetype(profile?.archetype),
		traits: profile?.traits ? profile.traits.map(normalizeTrait) : [...DEFAULT_TRAITS]
	};
}

export function profileFromDraft(draft: Draft): ProfilePayload {
	const name = draft.name.trim() || DEFAULT_NAME;
	return {
		name,
		avatar_url: avatarUrlFor(name),
		archetype: draft.archetype,
		traits: draft.traits.map(normalizeTrait)
	};
}

export function archetypeTint(archetype: string): readonly [string, string] {
	return TINTS[normalizeArchetype(archetype)];
}
