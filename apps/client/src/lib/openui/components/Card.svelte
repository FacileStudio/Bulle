<script lang="ts">
	import { Card as MuseCard } from '@facile/muse';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	let {
		props,
		renderNode
	}: ComponentRenderProps<{
		children?: unknown;
		title?: string | null;
		description?: string | null;
	}> = $props();
</script>

<MuseCard class="flex flex-col gap-4">
	{#if props.title || props.description}
		<div class="flex flex-col gap-1">
			{#if props.title}
				<h3 class="text-fc-md font-semibold text-fc-fg">{props.title}</h3>
			{/if}
			{#if props.description}
				<p class="text-fc-sm text-fc-fg-muted">{props.description}</p>
			{/if}
		</div>
	{/if}
	{@render renderNode(props.children)}
</MuseCard>
