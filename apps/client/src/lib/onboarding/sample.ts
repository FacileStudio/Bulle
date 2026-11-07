import { t } from '$lib/i18n/index.svelte';
import type { AppKey } from '$lib/i18n/messages';
import { ARCHETYPES, TRAITS } from '../assistant';
import { normalizeArchetype, normalizeTrait } from './profile';

export function samplePrompt(): string {
	return t('onboarding.samplePrompt');
}

type Voice = {
	opener: string;
	body: readonly string[];
	short: readonly string[];
	closer: string;
};

const VOICE_SIZES: Record<string, { body: number; short: number }> = {
	pragmatic_senior: { body: 3, short: 3 },
	terse_minimalist: { body: 1, short: 1 },
	supportive_mentor: { body: 3, short: 3 },
	witty_hacker: { body: 3, short: 3 }
};

function lines(id: string, part: 'body' | 'short', count: number): string[] {
	return Array.from({ length: count }, (_, index) => t(`onboarding.voice.${id}.${part}${index + 1}` as AppKey));
}

function voiceOf(id: string): Voice {
	const size = VOICE_SIZES[id];
	return {
		opener: t(`onboarding.voice.${id}.opener` as AppKey),
		body: lines(id, 'body', size.body),
		short: lines(id, 'short', size.short),
		closer: t(`onboarding.voice.${id}.closer` as AppKey)
	};
}

const TRAIT_ORDER = ['explain_reasoning', 'witty', 'strict', 'creative'] as const;

export function describeProfile(archetype: string, traits: readonly string[]): string {
	const label = ARCHETYPES.find((a) => a.id === normalizeArchetype(archetype))?.label ?? '';
	const names = traits.map(normalizeTrait).map((id) => TRAITS.find((t) => t.id === id)?.label ?? id);
	return [label, ...names.map((n) => n.toLowerCase())].join(', ');
}

export function composeSample(archetype: string, traits: readonly string[]): string[] {
	const voice = voiceOf(normalizeArchetype(archetype));
	const active = new Set(traits.map(normalizeTrait));
	const concise = active.has('concise');
	const out = concise ? [...voice.short] : [voice.opener, ...voice.body];
	for (const id of TRAIT_ORDER) {
		if (active.has(id)) out.push(t(`onboarding.traitLine.${id}` as AppKey));
	}
	if (!concise) out.push(voice.closer);
	return out;
}
