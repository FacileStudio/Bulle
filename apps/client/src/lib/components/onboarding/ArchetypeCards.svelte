<script lang="ts">
	import { Icon, icons } from '@facile/muse';
	import { ARCHETYPES } from '$lib/assistant';
	import { composeSample } from '$lib/onboarding/sample';

	let {
		archetype = $bindable(''),
		traits,
		showSample = true
	}: { archetype?: string; traits: readonly string[]; showSample?: boolean } = $props();
</script>

<div role="radiogroup" aria-label="Personality" class="grid w-full gap-3 sm:grid-cols-2">
	{#each ARCHETYPES as item (item.id)}
		<label
			class="relative flex cursor-pointer flex-col gap-3 rounded-fc-lg border border-fc-border bg-fc-bg/80 p-4 text-left transition-colors has-checked:border-fc-accent has-checked:ring-2 has-checked:ring-fc-accent has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-fc-ring"
		>
			<input
				type="radio"
				name="archetype"
				value={item.id}
				bind:group={archetype}
				class="sr-only"
			/>
			<span class="flex items-start justify-between gap-3">
				<span class="flex flex-col gap-0.5">
					<span class="text-fc-md font-medium text-fc-fg">{item.label}</span>
					<span class="text-fc-xs text-fc-fg-muted">{item.description}</span>
				</span>
				<Icon
					icon={icons.check}
					size={20}
					class="shrink-0 text-fc-accent transition-opacity {archetype === item.id
						? 'opacity-100'
						: 'opacity-0'}"
					aria-hidden="true"
				/>
			</span>
			{#if showSample}
				<span class="block">
					<span class="mb-1.5 block text-fc-xs font-medium text-fc-fg-muted">
						Sample reply
					</span>
					<span class="block rounded-fc-md bg-fc-surface p-3 text-fc-xs text-fc-fg">
						{#each composeSample(item.id, traits) as line, index (index)}
							<span class="block {index > 0 ? 'mt-1.5' : ''}">{line}</span>
						{/each}
					</span>
				</span>
			{/if}
		</label>
	{/each}
</div>
