import { ARCHETYPES, TRAITS } from '../assistant';
import { normalizeArchetype, normalizeTrait } from './profile';

export const SAMPLE_PROMPT = "What's on my plate today?";

type Voice = {
	opener: string;
	body: readonly string[];
	short: readonly string[];
	closer: string;
};

const VOICES: Record<string, Voice> = {
	pragmatic_senior: {
		opener: 'Three things matter today.',
		body: [
			'1. A client email about invoice 42. It blocks payment, so reply first.',
			'2. Standup at 10:00, design review at 14:30.',
			'3. The API timeout bug is due today. Budget two hours.'
		],
		short: [
			'Invoice 42 reply blocks payment. Do it first.',
			'Standup 10:00, review 14:30.',
			'API timeout bug is due today.'
		],
		closer: 'Start with the invoice. Everything else fits around it.'
	},
	terse_minimalist: {
		opener: 'Today:',
		body: ['Invoice 42 reply. Standup 10:00. Review 14:30. API timeout bug.'],
		short: ['Invoice 42 reply. Standup 10:00. Review 14:30. API timeout bug.'],
		closer: 'Invoice first.'
	},
	supportive_mentor: {
		opener: 'Good morning. Here is how your day looks and where I would put your energy.',
		body: [
			'A client wrote about invoice 42. It is the one thing waiting on you, so answering it first will clear your head.',
			'Standup is at 10:00 and the design review is at 14:30, which leaves a long quiet stretch between them.',
			'The API timeout bug is due today, and that quiet stretch is a good place for it.'
		],
		short: [
			'Start with the client email about invoice 42.',
			'Standup is at 10:00 and the review at 14:30.',
			'The API timeout bug fits in the gap between them.'
		],
		closer: 'You have enough room today. Take it one step at a time.'
	},
	witty_hacker: {
		opener: 'Your day, compiled with zero warnings.',
		body: [
			'Invoice 42: the client wants a reply. Money does not move itself.',
			'Standup at 10:00 and a design review at 14:30. Two meetings, one of which might have survived as a message.',
			'The API timeout bug is due today. It has been timing out right on schedule, which is at least consistent.'
		],
		short: [
			'Invoice 42 needs a reply.',
			'Standup 10:00, review 14:30.',
			'The API timeout bug is due today.'
		],
		closer: 'Reply to the invoice, then go hunting.'
	}
};

const TRAIT_LINES: Record<string, string> = {
	explain_reasoning:
		'Why this order: the invoice blocks payment, the meetings are fixed, and the bug is the only item you can move.',
	witty: 'The review is at 14:30 again. Bring snacks.',
	strict:
		'For the bug I will add a type that rejects a missing timeout, so the compiler catches it next time.',
	creative: 'Idea: batch the two replies after standup and spend the free hour on the bug.'
};

const TRAIT_ORDER = ['explain_reasoning', 'witty', 'strict', 'creative'] as const;

export function describeProfile(archetype: string, traits: readonly string[]): string {
	const label = ARCHETYPES.find((a) => a.id === normalizeArchetype(archetype))?.label ?? '';
	const names = traits.map(normalizeTrait).map((id) => TRAITS.find((t) => t.id === id)?.label ?? id);
	return [label, ...names.map((n) => n.toLowerCase())].join(', ');
}

export function composeSample(archetype: string, traits: readonly string[]): string[] {
	const voice = VOICES[normalizeArchetype(archetype)];
	const active = new Set(traits.map(normalizeTrait));
	const concise = active.has('concise');
	const lines = concise ? [...voice.short] : [voice.opener, ...voice.body];
	for (const id of TRAIT_ORDER) {
		if (active.has(id)) lines.push(TRAIT_LINES[id]);
	}
	if (!concise) lines.push(voice.closer);
	return lines;
}
