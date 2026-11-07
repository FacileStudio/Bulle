<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { TRAITS } from '$lib/assistant';
	import { hasTrait, toggleTrait } from '$lib/onboarding/profile';

	let {
		traits = $bindable([]),
		class: className = ''
	}: { traits?: string[]; class?: string } = $props();
</script>

<div class="flex flex-wrap gap-2 {className}" role="group" aria-label={t('onboarding.traits')}>
	{#each TRAITS as item (item.id)}
		{@const active = hasTrait(traits, item.id)}
		<button
			type="button"
			aria-pressed={active}
			onclick={() => (traits = toggleTrait(traits, item.id))}
			class="inline-flex min-h-11 cursor-pointer items-center rounded-fc-pill px-4 text-fc-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring {active
				? 'bg-fc-accent text-fc-accent-fg'
				: 'bg-fc-component text-fc-fg'}"
		>
			{item.label}
		</button>
	{/each}
</div>
