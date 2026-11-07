<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { STEPS, progressText } from '$lib/onboarding/steps';

	let { index }: { index: number } = $props();
</script>

<div class="flex w-full max-w-fc-sm flex-col gap-2">
	<div class="flex items-center justify-between text-fc-xs text-fc-fg-muted">
		<span>{t('onboarding.stepOf', { step: index + 1, total: STEPS.length })}</span>
		<span>{STEPS[index].label}</span>
	</div>
	<div
		role="progressbar"
		aria-label={t('onboarding.progress')}
		aria-valuemin={1}
		aria-valuemax={STEPS.length}
		aria-valuenow={index + 1}
		aria-valuetext={progressText(index)}
		class="flex gap-1.5"
	>
		{#each STEPS as item, position (item.id)}
			<span
				class="h-1.5 flex-1 rounded-fc-pill transition-colors duration-500 {position <= index
					? 'bg-fc-accent'
					: 'bg-fc-surface'}"
			></span>
		{/each}
	</div>
</div>
