<script lang="ts">
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	let { props }: ComponentRenderProps<{ items?: string[]; ordered?: boolean | null }> = $props();

	const items = $derived(props.items ?? []);
	const classes = 'flex flex-col gap-1 pl-5 text-fc-sm text-fc-fg';
</script>

{#if props.ordered}
	<ol class="{classes} list-decimal">
		{#each items as item}<li>{item}</li>{/each}
	</ol>
{:else}
	<ul class="{classes} list-disc">
		{#each items as item}<li>{item}</li>{/each}
	</ul>
{/if}
