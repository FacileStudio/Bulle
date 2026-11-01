<script lang="ts">
	import { BarChart, LineChart } from '@facile/muse';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	let {
		props
	}: ComponentRenderProps<{
		kind?: 'bar' | 'line';
		labels?: string[];
		series?: { name?: string; data?: number[] }[];
		title?: string | null;
	}> = $props();

	const labels = $derived(props.labels ?? []);
	const series = $derived(
		(props.series ?? []).map((s) => ({ name: s.name ?? '', data: s.data ?? [] }))
	);
</script>

<figure class="flex flex-col gap-2">
	{#if props.title}
		<figcaption class="text-fc-sm font-medium text-fc-fg">{props.title}</figcaption>
	{/if}
	{#if props.kind === 'line'}
		<LineChart {series} {labels} height={220} />
	{:else}
		<BarChart {series} {labels} height={220} />
	{/if}
</figure>
