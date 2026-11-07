<script lang="ts">
	import { richInline } from './MarkdownMuseRich';
	import {
		parseBarChart,
		parseCompare,
		parseListItem,
		parseMetrics,
		type CompareSection,
		type FenceKind
	} from './MarkdownMuseUtils';

	let { kind, source }: { kind: Exclude<FenceKind, 'mermaid' | 'code'>; source: string } = $props();

	const chart = $derived(kind === 'bars' ? parseBarChart(source) : null);
	const metrics = $derived(kind === 'metrics' ? parseMetrics(source) : []);
	const sections = $derived(kind === 'compare' ? parseCompare(source) : []);

	const SECTION_TONES: Record<CompareSection['tone'], { box: string; text: string; dot: string; mark: string }> = {
		pro: { box: 'bg-fc-success/10', text: 'text-fc-success', dot: 'bg-fc-success', mark: '+' },
		con: { box: 'bg-fc-danger/10', text: 'text-fc-danger', dot: 'bg-fc-danger', mark: '-' },
		neutral: { box: 'bg-fc-surface', text: 'text-fc-fg', dot: 'bg-fc-accent', mark: '*' }
	};
</script>

{#if chart}
	<figure class="flex flex-col gap-4 rounded-fc-lg bg-fc-surface p-5">
		{#if chart.title}
			<figcaption class="text-fc-sm font-semibold text-fc-fg">{chart.title}</figcaption>
		{/if}
		<div class="flex flex-col gap-3.5">
			{#each chart.items as item, i (i)}
				{@const pct = Math.min(Math.round((item.value / chart.max) * 100), 100)}
				<div class="flex flex-col gap-1 text-fc-xs">
					<div class="flex justify-between gap-3 font-medium text-fc-fg">
						<span class="min-w-0">{item.label}</span>
						<span class="shrink-0 font-fc-mono text-fc-fg-muted">{item.formatted}</span>
					</div>
					<div class="h-2 w-full overflow-hidden rounded-full bg-fc-component" aria-hidden="true">
						<div class="h-full rounded-full bg-fc-accent" style:width="{pct}%"></div>
					</div>
				</div>
			{/each}
		</div>
	</figure>
{:else if kind === 'metrics'}
	<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
		{#each metrics as m, i (i)}
			<div class="flex flex-col gap-1 rounded-fc-lg bg-fc-surface p-4">
				{#if m.label}
					<span class="text-fc-xs font-semibold tracking-wider text-fc-fg-muted uppercase">{m.label}</span>
				{/if}
				{#if m.val}
					<span class="text-fc-xl font-bold tracking-tight text-fc-fg">{m.val}</span>
				{/if}
				{#if m.sub}
					<span class="text-fc-xs text-fc-fg-muted">{m.sub}</span>
				{/if}
			</div>
		{/each}
	</div>
{:else}
	<div class="grid gap-4 sm:grid-cols-2">
		{#each sections as section, i (i)}
			{@const tone = SECTION_TONES[section.tone]}
			<div class="flex flex-col gap-3 rounded-fc-lg p-4 {tone.box}">
				<div class="flex items-center gap-2 font-semibold {tone.text}">
					<span class="inline-flex size-5 items-center justify-center rounded-full bg-fc-component text-fc-xs font-bold" aria-hidden="true"
						>{tone.mark}</span
					>
					<span>{section.title}</span>
				</div>
				<ul class="flex flex-col gap-2 text-fc-sm text-fc-fg-muted">
					{#each section.items as entry, j (j)}
						<li class="flex items-start gap-2">
							<span class="mt-2 size-1 shrink-0 rounded-full {tone.dot}" aria-hidden="true"></span>
							<span class="min-w-0">{@html richInline(parseListItem(entry.replace(/^[-*]\s+/, '')).clean)}</span>
						</li>
					{/each}
				</ul>
			</div>
		{/each}
	</div>
{/if}
