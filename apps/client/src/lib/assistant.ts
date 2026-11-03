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
		label: 'Pragmatic Senior',
		description: 'Direct, battle-tested, clean code first.'
	},
	{
		id: 'terse_minimalist',
		label: 'Terse Minimalist',
		description: 'Ultra concise, code-heavy, minimum words.'
	},
	{
		id: 'supportive_mentor',
		label: 'Supportive Mentor',
		description: 'Thorough, clear explanations, patient and insightful.'
	},
	{
		id: 'witty_hacker',
		label: 'Witty Hacker',
		description: 'Clever, lightly sarcastic, highly technical, loves elegant hacks.'
	}
] as const;

export const TRAITS: readonly TraitItem[] = [
	{ id: 'concise', label: 'Concise' },
	{ id: 'explain_reasoning', label: 'Explain Reasoning' },
	{ id: 'witty', label: 'Light Sarcasm' },
	{ id: 'strict', label: 'Strict Type Safety' },
	{ id: 'creative', label: 'Creative Solutions' }
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
