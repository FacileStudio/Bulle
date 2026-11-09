<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		body,
		badge,
		children
	}: {
		title: string;
		body: string;
		badge?: Snippet;
		children: Snippet;
	} = $props();
</script>

<article class="flex flex-col overflow-hidden rounded-fc-lg bg-fc-component">
	<div class="relative h-60 overflow-hidden" aria-hidden="true" inert>
		{@render children()}
	</div>
	<div class="flex flex-col gap-1.5 px-6 pt-2 pb-6">
		<div class="flex items-center justify-between gap-2">
			<h3 class="font-fc-title text-fc-lg font-semibold text-fc-fg">{title}</h3>
			{#if badge}
				{@render badge()}
			{/if}
		</div>
		<p class="text-fc-sm text-fc-fg-muted">{body}</p>
	</div>
</article>
