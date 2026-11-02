<script lang="ts">
	import { goto } from '$app/navigation';
	import { tick } from 'svelte';
	import { Alert } from '@facile/muse';
	import { updateAssistant } from '$lib/backend';
	import type { CaretCoordinates } from '$lib/caret';
	import Logo from '$lib/components/Logo.svelte';
	import SuiteConnections from '$lib/components/SuiteConnections.svelte';
	import { popAvatar, revealStep } from '$lib/onboarding/motion';
	import { DEFAULT_NAME, archetypeTint, initialDraft, profileFromDraft } from '$lib/onboarding/profile';
	import { SAMPLE_PROMPT, composeSample, describeProfile } from '$lib/onboarding/sample';
	import {
		STEPS,
		advance,
		applySkip,
		isLastStep,
		primaryLabel,
		progressText,
		retreat
	} from '$lib/onboarding/steps';
	import { clearDraft, loadDraft, saveDraft } from '$lib/onboarding/storage';
	import { session, setAssistant } from '$lib/session.svelte';
	import Ambient from './Ambient.svelte';
	import ArchetypeCards from './ArchetypeCards.svelte';
	import NameField from './NameField.svelte';
	import OnboardingAvatar from './OnboardingAvatar.svelte';
	import SampleReply from './SampleReply.svelte';
	import StepHeading from './StepHeading.svelte';
	import StepNav from './StepNav.svelte';
	import StepProgress from './StepProgress.svelte';
	import TraitChips from './TraitChips.svelte';

	const AVATAR_SIZES = { welcome: 176, name: 144, personality: 72, traits: 72, apps: 72, meet: 200 } as const;

	let draft = $state(loadDraft() ?? initialDraft(session.user?.assistant));
	let direction = $state<1 | -1>(1);
	let saving = $state(false);
	let error = $state<string | null>(null);
	let caret = $state<CaretCoordinates | null>(null);
	let content = $state<HTMLElement | null>(null);
	let avatarBox = $state<HTMLElement | null>(null);

	const step = $derived(STEPS[draft.step]);
	const displayName = $derived(draft.name.trim() || DEFAULT_NAME);
	const sample = $derived(composeSample(draft.archetype, draft.traits));

	$effect(() => {
		saveDraft($state.snapshot(draft));
	});

	$effect(() => {
		const current = step.id;
		void tick().then(() => {
			revealStep(content, direction);
			popAvatar(avatarBox, current === 'meet' || current === 'welcome');
			content?.querySelector<HTMLElement>('[data-step-heading]')?.focus({ preventScroll: true });
		});
	});

	function go(index: number, dir: 1 | -1) {
		direction = dir;
		caret = null;
		error = null;
		draft.step = index;
	}

	async function finish() {
		if (saving) return;
		saving = true;
		error = null;
		try {
			setAssistant(await updateAssistant(profileFromDraft($state.snapshot(draft))));
			clearDraft();
			await goto('/sessions', { replaceState: true });
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not save your assistant';
		} finally {
			saving = false;
		}
	}

	function onContinue() {
		if (isLastStep(draft.step)) void finish();
		else go(advance(draft.step), 1);
	}

	function onSkip() {
		direction = 1;
		caret = null;
		draft = applySkip($state.snapshot(draft));
	}

	function onKeydown(event: KeyboardEvent) {
		if (event.key !== 'Enter' || event.isComposing) return;
		const target = event.target;
		const fromField = target instanceof HTMLInputElement;
		const fromHeading = target instanceof HTMLElement && target.hasAttribute('data-step-heading');
		if (!fromField && !fromHeading) return;
		event.preventDefault();
		onContinue();
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div class="fixed inset-0 z-50 flex h-dvh flex-col overflow-hidden bg-fc-page text-fc-fg">
	<Ambient tint={archetypeTint(draft.archetype)} />

	<header class="relative z-10 flex flex-col items-center gap-4 px-4 pt-4 sm:px-8 sm:pt-6">
		<div class="flex items-center gap-2 text-fc-fg-muted">
			<Logo size={20} />
			<span class="text-fc-sm font-medium">Bulle</span>
		</div>
		<StepProgress index={draft.step} />
	</header>

	<main class="relative z-10 min-h-0 flex-1 overflow-y-auto overscroll-contain">
		<div
			class="mx-auto flex min-h-full w-full max-w-fc-lg flex-col items-center justify-center gap-6 px-4 py-6 text-center"
		>
			<div bind:this={avatarBox}>
				<OnboardingAvatar name={displayName} size={AVATAR_SIZES[step.id]} {caret} />
			</div>

			<div bind:this={content} class="flex w-full flex-col items-center gap-6">
				{#if step.id === 'welcome'}
					<StepHeading
						title="Meet your new assistant."
						description="It handles your mail, calendar and tasks, and writes code on its own computer. Setup takes about a minute."
					/>
				{:else if step.id === 'name'}
					<StepHeading
						title="What should it be called?"
						description="Pick any name. Its avatar is drawn from it."
					/>
					<div class="w-full max-w-fc-sm" data-reveal>
						<NameField
							id="onboarding-name"
							bind:value={draft.name}
							hideLabel
							oncaret={(coords) => (caret = coords)}
						/>
					</div>
				{:else if step.id === 'personality'}
					<StepHeading
						title={`How should ${displayName} talk?`}
						description={`Four voices answering the same question: ${SAMPLE_PROMPT}`}
					/>
					<div class="w-full" data-reveal>
						<ArchetypeCards bind:archetype={draft.archetype} traits={draft.traits} />
					</div>
				{:else if step.id === 'traits'}
					<StepHeading title="Tune the style" />
					<div class="w-full" data-reveal>
						<TraitChips bind:traits={draft.traits} />
					</div>
					<div class="w-full max-w-fc-sm" data-reveal>
						<SampleReply lines={sample} live />
					</div>
				{:else if step.id === 'apps'}
					<StepHeading
						title="Connect your apps"
						description={`${displayName} reads your own mail, calendar and tasks. You can skip this and open an app later.`}
					/>
					<div class="w-full max-w-fc-sm text-left" data-reveal>
						<SuiteConnections />
					</div>
				{:else}
					<StepHeading
						title={`Meet ${displayName}`}
						description={describeProfile(draft.archetype, draft.traits)}
					/>
					{#if error}
						<div class="w-full max-w-fc-sm" data-reveal>
							<Alert tone="danger">{error}</Alert>
						</div>
					{/if}
				{/if}
			</div>
		</div>
	</main>

	<StepNav
		primary={primaryLabel(draft.step, draft.name)}
		busy={saving}
		final={isLastStep(draft.step)}
		canBack={draft.step > 0}
		canSkip={step.skippable}
		oncontinue={onContinue}
		onback={() => go(retreat(draft.step), -1)}
		onskip={onSkip}
	/>

	<p class="sr-only" aria-live="polite">{progressText(draft.step)}</p>
</div>
