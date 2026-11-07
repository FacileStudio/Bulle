import { t } from '$lib/i18n/index.svelte';
import { DEFAULT_NAME, DEFAULT_TRAITS, type Draft } from './profile';

export type StepId = 'welcome' | 'name' | 'personality' | 'traits' | 'apps' | 'meet';

export type Step = {
	id: StepId;
	label: string;
	skippable: boolean;
};

export const STEPS: readonly Step[] = [
	{
		id: 'welcome',
		get label() {
			return t('onboarding.step.welcome');
		},
		skippable: false
	},
	{
		id: 'name',
		get label() {
			return t('onboarding.step.name');
		},
		skippable: true
	},
	{
		id: 'personality',
		get label() {
			return t('onboarding.step.personality');
		},
		skippable: false
	},
	{
		id: 'traits',
		get label() {
			return t('onboarding.step.traits');
		},
		skippable: true
	},
	{
		id: 'apps',
		get label() {
			return t('onboarding.step.apps');
		},
		skippable: true
	},
	{
		id: 'meet',
		get label() {
			return t('onboarding.step.meet');
		},
		skippable: false
	}
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
			return t('onboarding.primary.welcome');
		case 'name':
			return t('onboarding.primary.name');
		case 'personality':
			return t('onboarding.primary.personality');
		case 'traits':
			return t('onboarding.primary.traits');
		case 'apps':
			return t('onboarding.primary.apps', { name: name.trim() || DEFAULT_NAME });
		default:
			return t('onboarding.primary.meet');
	}
}

export function needsOnboarding(
	user: { assistant?: { name?: string | null } | null } | null
): boolean {
	return !user?.assistant?.name?.trim();
}

export function progressText(index: number): string {
	const step = clampStep(index);
	return t('onboarding.progressText', { step: step + 1, total: STEPS.length, label: STEPS[step].label });
}
