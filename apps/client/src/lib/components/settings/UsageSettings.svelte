<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, BarChart, Section, Skeleton, StatCard } from '@facile/muse';
	import { getUsage, type UsageGroup, type UsageSummary, type UsageTotals } from '$lib/backend';
	import { formatCost, formatTokens } from '$lib/chat/format';
	import { formatDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	let summary = $state<UsageSummary | null>(null);
	let error = $state<string | null>(null);

	const windows = $derived(
		summary
			? [
					{ label: t('usage.today'), totals: summary.today },
					{ label: t('usage.week'), totals: summary.week },
					{ label: t('usage.month'), totals: summary.month }
				]
			: []
	);
	const labels = $derived(
		summary?.days.map((d) =>
			formatDate(`${d.date}T12:00:00`, { day: 'numeric', month: 'short' })
		) ?? []
	);
	const series = $derived(
		summary ? [{ name: t('usage.cost'), data: summary.days.map((d) => Number(d.cost_usd.toFixed(4))) }] : []
	);
	const month = $derived(summary?.month);

	function tokens(totals: UsageTotals): number {
		return (
			totals.input_tokens +
			totals.output_tokens +
			totals.cache_read_tokens +
			totals.cache_creation_tokens
		);
	}

	function tokenText(n: number): string {
		return t('usage.tokens', { count: formatTokens(n) });
	}

	function runs(n: number): string {
		return t('usage.runs', { count: n });
	}

	onMount(async () => {
		try {
			summary = await getUsage();
		} catch (err) {
			error = err instanceof Error ? err.message : t('usage.loadFailed');
		}
	});
</script>

{#snippet breakdown(title: string, groups: UsageGroup[])}
	<div class="flex flex-col gap-2">
		<h3 class="text-fc-sm font-medium">{title}</h3>
		{#if groups.length}
			<ul class="flex flex-col text-fc-sm">
				{#each groups as group (group.name)}
					<li class="flex items-center justify-between gap-3 py-2">
						<span class="min-w-0 truncate">{group.name}</span>
						<span class="shrink-0 text-fc-fg-muted">
							{runs(group.runs)} &middot; {tokenText(tokens(group))} &middot;
							<span class="text-fc-fg">{formatCost(group.cost_usd)}</span>
						</span>
					</li>
				{/each}
			</ul>
		{:else}
			<p class="text-fc-sm text-fc-fg-muted">{t('usage.nothing')}</p>
		{/if}
	</div>
{/snippet}

<Section card id="usage" title={t('usage.title')} description={t('usage.description')}>
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
						delta={t('usage.delta', { tokens: tokenText(tokens(w.totals)), runs: runs(w.totals.runs) })}
					/>
				{/each}
			</div>

			<div class="flex flex-col gap-2">
				<h3 class="text-fc-sm font-medium">{t('usage.daily')}</h3>
				<BarChart {series} {labels} height={180} yFormat={formatCost} emptyLabel={t('usage.noRuns')} />
			</div>

			<dl class="grid grid-cols-2 gap-x-4 gap-y-2 text-fc-sm sm:grid-cols-4">
				<div>
					<dt class="text-fc-fg-muted">{t('usage.input')}</dt>
					<dd>{formatTokens(month.input_tokens)}</dd>
				</div>
				<div>
					<dt class="text-fc-fg-muted">{t('usage.output')}</dt>
					<dd>{formatTokens(month.output_tokens)}</dd>
				</div>
				<div>
					<dt class="text-fc-fg-muted">{t('usage.cacheRead')}</dt>
					<dd>{formatTokens(month.cache_read_tokens)}</dd>
				</div>
				<div>
					<dt class="text-fc-fg-muted">{t('usage.cacheWrite')}</dt>
					<dd>{formatTokens(month.cache_creation_tokens)}</dd>
				</div>
			</dl>

			<div class="grid gap-6 sm:grid-cols-2">
				{@render breakdown(t('usage.byModel'), summary.models)}
				{@render breakdown(t('usage.byChannel'), summary.channels)}
			</div>
			<p class="text-fc-xs text-fc-fg-muted">{t('usage.footer', { timezone: summary.timezone })}</p>
		{/if}
	</div>
</Section>
