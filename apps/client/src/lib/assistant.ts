import { t } from '$lib/i18n/index.svelte';
import { blobUri } from '@facile/blob';

export type ArchetypeItem = {
	id: string;
	label: string;
	description: string;
};

export type TraitItem = {
	id: string;
	label: string;
};

export const ARCHETYPES: readonly ArchetypeItem[] = [
	{
		id: 'pragmatic_senior',
		get label() {
			return t('assistant.archetype.pragmatic_senior.label');
		},
		get description() {
			return t('assistant.archetype.pragmatic_senior.description');
		}
	},
	{
		id: 'terse_minimalist',
		get label() {
			return t('assistant.archetype.terse_minimalist.label');
		},
		get description() {
			return t('assistant.archetype.terse_minimalist.description');
		}
	},
	{
		id: 'supportive_mentor',
		get label() {
			return t('assistant.archetype.supportive_mentor.label');
		},
		get description() {
			return t('assistant.archetype.supportive_mentor.description');
		}
	},
	{
		id: 'witty_hacker',
		get label() {
			return t('assistant.archetype.witty_hacker.label');
		},
		get description() {
			return t('assistant.archetype.witty_hacker.description');
		}
	}
] as const;

export const TRAITS: readonly TraitItem[] = [
	{
		id: 'concise',
		get label() {
			return t('assistant.trait.concise');
		}
	},
	{
		id: 'explain_reasoning',
		get label() {
			return t('assistant.trait.explain_reasoning');
		}
	},
	{
		id: 'witty',
		get label() {
			return t('assistant.trait.witty');
		}
	},
	{
		id: 'strict',
		get label() {
			return t('assistant.trait.strict');
		}
	},
	{
		id: 'creative',
		get label() {
			return t('assistant.trait.creative');
		}
	}
] as const;

export const RANDOM_NAMES = [
	'Ada',
	'Atlas',
	'Echo',
	'Jarvis',
	'Nova',
	'Pixel',
	'Sage',
	'Bolt',
	'Cipher',
	'Kori',
	'Orion',
	'Cosmo',
	'Freya',
	'Pascal',
	'Turing',
	'Lovelace'
] as const;

export function avatarUrlFor(name: string): string {
	return blobUri(name.trim() || 'bulle');
}
