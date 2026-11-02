import { DEFAULT_NAME, DEFAULT_TRAITS, type Draft } from './profile';

export type StepId = 'welcome' | 'name' | 'personality' | 'traits' | 'apps' | 'meet';

export type Step = {
	id: StepId;
	label: string;
	skippable: boolean;
};

export const STEPS: readonly Step[] = [
	{ id: 'welcome', label: 'Welcome', skippable: false },
	{ id: 'name', label: 'Name', skippable: true },
	{ id: 'personality', label: 'Personality', skippable: false },
	{ id: 'traits', label: 'Style', skippable: true },
	{ id: 'apps', label: 'Apps', skippable: true },
	{ id: 'meet', label: 'Meet', skippable: false }
];

const LAST = STEPS.length - 1;

export function clampStep(index: number): number {
	if (!Number.isFinite(index)) return 0;
	return Math.min(LAST, Math.max(0, Math.trunc(index)));
}

export function advance(index: number): number {
	return clampStep(index + 1);
}

export function retreat(index: number): number {
	return clampStep(index - 1);
}

export function isLastStep(index: number): boolean {
	return clampStep(index) === LAST;
}

export function applySkip(draft: Draft): Draft {
	const step = STEPS[clampStep(draft.step)];
	if (!step.skippable) return draft;
	const cleared =
		step.id === 'name'
			? { name: '' }
			: step.id === 'traits'
				? { traits: [...DEFAULT_TRAITS] }
				: {};
	return { ...draft, ...cleared, step: advance(draft.step) };
}

export function primaryLabel(index: number, name: string): string {
	switch (STEPS[clampStep(index)].id) {
		case 'welcome':
			return 'Give it a name';
		case 'name':
			return 'Choose a personality';
		case 'personality':
			return 'Fine-tune the style';
		case 'traits':
			return 'Connect your apps';
		case 'apps':
			return `Meet ${name.trim() || DEFAULT_NAME}`;
		default:
			return 'Say hello';
	}
}

export function needsOnboarding(
	user: { assistant?: { name?: string | null } | null } | null
): boolean {
	return !user?.assistant?.name?.trim();
}

export function progressText(index: number): string {
	const step = clampStep(index);
	return `Step ${step + 1} of ${STEPS.length}: ${STEPS[step].label}`;
}
