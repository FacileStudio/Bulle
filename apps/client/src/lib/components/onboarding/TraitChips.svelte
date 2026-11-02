<script lang="ts">
	import { TRAITS } from '$lib/assistant';
	import { hasTrait, toggleTrait } from '$lib/onboarding/profile';

	let { traits = $bindable([]) }: { traits?: string[] } = $props();
</script>

<div class="flex flex-wrap justify-center gap-2" role="group" aria-label="Traits">
	{#each TRAITS as item (item.id)}
		{@const active = hasTrait(traits, item.id)}
		<button
			type="button"
			aria-pressed={active}
			onclick={() => (traits = toggleTrait(traits, item.id))}
			class="inline-flex min-h-11 cursor-pointer items-center rounded-fc-pill border px-4 text-fc-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring {active
				? 'border-fc-accent bg-fc-accent text-fc-accent-fg'
				: 'border-fc-border bg-fc-bg/80 text-fc-fg'}"
		>
			{item.label}
		</button>
	{/each}
</div>
