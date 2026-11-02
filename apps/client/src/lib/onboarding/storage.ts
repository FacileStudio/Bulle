import { normalizeArchetype, normalizeTrait, NAME_MAX_LENGTH, type Draft } from './profile';
import { clampStep } from './steps';

export const DRAFT_KEY = 'bulle.onboarding.draft';

export type DraftStore = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>;

function defaultStore(): DraftStore | null {
	try {
		return globalThis.sessionStorage ?? null;
	} catch {
		return null;
	}
}

function parseDraft(raw: string): Draft | null {
	const value: unknown = JSON.parse(raw);
	if (typeof value !== 'object' || value === null) return null;
	const { step, name, archetype, traits } = value as Record<string, unknown>;
	if (typeof name !== 'string' || !Array.isArray(traits)) return null;
	return {
		step: clampStep(typeof step === 'number' ? step : 0),
		name: name.slice(0, NAME_MAX_LENGTH),
		archetype: normalizeArchetype(typeof archetype === 'string' ? archetype : null),
		traits: traits.filter((t): t is string => typeof t === 'string').map(normalizeTrait)
	};
}

export function loadDraft(store: DraftStore | null = defaultStore()): Draft | null {
	try {
		const raw = store?.getItem(DRAFT_KEY);
		return raw ? parseDraft(raw) : null;
	} catch {
		return null;
	}
}

export function saveDraft(draft: Draft, store: DraftStore | null = defaultStore()): void {
	try {
		store?.setItem(DRAFT_KEY, JSON.stringify(draft));
	} catch {
		return;
	}
}

export function clearDraft(store: DraftStore | null = defaultStore()): void {
	try {
		store?.removeItem(DRAFT_KEY);
	} catch {
		return;
	}
}
