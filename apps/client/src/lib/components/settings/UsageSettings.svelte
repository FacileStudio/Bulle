<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, BarChart, Section, Skeleton, StatCard } from '@facile/muse';
	import { getUsage, type UsageGroup, type UsageSummary, type UsageTotals } from '$lib/backend';
	import { formatCost, formatTokens } from '$lib/chat/format';

	let summary = $state<UsageSummary | null>(null);
	let error = $state<string | null>(null);

	const windows = $derived(
		summary
			? [
					{ label: 'Today', totals: summary.today },
					{ label: 'Last 7 days', totals: summary.week },
					{ label: 'Last 30 days', totals: summary.month }
				]
			: []
	);
	const labels = $derived(
		summary?.days.map((d) =>
			new Date(`${d.date}T12:00:00`).toLocaleDateString(undefined, { day: 'numeric', month: 'short' })
		) ?? []
	);
	const series = $derived(
		summary ? [{ name: 'Cost', data: summary.days.map((d) => Number(d.cost_usd.toFixed(4))) }] : []
	);
	const month = $derived(summary?.month);

	function tokens(t: UsageTotals): number {
		return t.input_tokens + t.output_tokens + t.cache_read_tokens + t.cache_creation_tokens;
	}

	function runs(n: number): string {
		return `${n} ${n === 1 ? 'run' : 'runs'}`;
	}

	onMount(async () => {
		try {
			summary = await getUsage();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not load usage';
		}
	});
</script>

{#snippet breakdown(title: string, groups: UsageGroup[])}
	<div class="flex flex-col gap-2">
		<h3 class="text-fc-sm font-medium">{title}</h3>
		{#if groups.length}
			<ul class="flex flex-col divide-y divide-fc-border text-fc-sm">
				{#each groups as group (group.name)}
					<li class="flex items-center justify-between gap-3 py-2">
						<span class="min-w-0 truncate">{group.name}</span>
						<span class="shrink-0 text-fc-fg-muted">
							{runs(group.runs)} &middot; {formatTokens(tokens(group))} tokens &middot;
							<span class="text-fc-fg">{formatCost(group.cost_usd)}</span>
						</span>
					</li>
				{/each}
			</ul>
		{:else}
			<p class="text-fc-sm text-fc-fg-muted">Nothing in the last 30 days.</p>
		{/if}
	</div>
{/snippet}

<Section card id="usage" title="Usage" description="Tokens and cost of your conversations and background jobs.">
	<div class="flex flex-col gap-6">
		{#if error}
			<Alert tone="danger">{error}</Alert>
		{:else if !summary || !month}
			<Skeleton class="h-24" />
			<Skeleton class="h-48" />
		{:else}
			<div class="grid gap-4 sm:grid-cols-3">
				{#each windows as w (w.label)}
					<StatCard
						label={w.label}
						value={formatCost(w.totals.cost_usd)}
						delta={`${formatTokens(tokens(w.totals))} tokens · ${runs(w.totals.runs)}`}
					/>
				{/each}
			</div>

			<div class="flex flex-col gap-2">
				<h3 class="text-fc-sm font-medium">Daily cost</h3>
				<BarChart {series} {labels} height={180} yFormat={formatCost} emptyLabel="No runs yet" />
			</div>

			<dl class="grid grid-cols-2 gap-x-4 gap-y-2 text-fc-sm sm:grid-cols-4">
				<div>
					<dt class="text-fc-fg-muted">Input</dt>
					<dd>{formatTokens(month.input_tokens)}</dd>
				</div>
				<div>
					<dt class="text-fc-fg-muted">Output</dt>
					<dd>{formatTokens(month.output_tokens)}</dd>
				</div>
				<div>
					<dt class="text-fc-fg-muted">Cache read</dt>
					<dd>{formatTokens(month.cache_read_tokens)}</dd>
				</div>
				<div>
					<dt class="text-fc-fg-muted">Cache write</dt>
					<dd>{formatTokens(month.cache_creation_tokens)}</dd>
				</div>
			</dl>

			<div class="grid gap-6 sm:grid-cols-2">
				{@render breakdown('By model', summary.models)}
				{@render breakdown('Chat and jobs', summary.channels)}
			</div>
			<p class="text-fc-xs text-fc-fg-muted">Last 30 days, in {summary.timezone} time.</p>
		{/if}
	</div>
</Section>
