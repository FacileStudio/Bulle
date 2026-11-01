<script lang="ts">
	import { Card as MuseCard } from '@facile/muse';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	let { props, renderNode }: ComponentRenderProps<{ children?: unknown; title?: string | null }> =
		$props();
</script>

<MuseCard class="flex flex-col gap-4">
	{#if props.title}
		<h3 class="text-fc-md font-semibold text-fc-fg">{props.title}</h3>
	{/if}
	{@render renderNode(props.children)}
</MuseCard>
