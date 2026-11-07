<script lang="ts">
	import { richBlock, richInline } from './MarkdownMuseRich';
	import type { ListKind } from './MarkdownMuseUtils';

	let { kind, clean }: { kind: ListKind; clean: string } = $props();

	const html = $derived(clean.includes('\n') ? richBlock(clean) : richInline(clean));

	const KINDS = {
		pro: { row: 'bg-fc-success/10', chip: 'bg-fc-success/20 text-fc-success', mark: '+' },
		con: { row: 'bg-fc-danger/10', chip: 'bg-fc-danger/20 text-fc-danger', mark: '-' },
		warn: { row: 'bg-fc-warning/10', chip: 'bg-fc-warning/20 text-fc-warning', mark: '!' },
		info: { row: 'bg-fc-info/10', chip: 'bg-fc-info/20 text-fc-info', mark: 'i' }
	} as const;
</script>

{#if kind === 'normal'}
	<li class="flex items-start gap-2.5 pl-2 text-fc-sm text-fc-fg-muted">
		<span class="mt-2 size-1.5 shrink-0 rounded-full bg-fc-accent/80" aria-hidden="true"></span>
		<div class="md-flow min-w-0 leading-relaxed">{@html html}</div>
	</li>
{:else}
	{@const k = KINDS[kind]}
	<li class="flex items-start gap-2.5 rounded-fc-md px-3 py-1.5 text-fc-sm {k.row}">
		<span
			class="mt-0.5 inline-flex size-4 shrink-0 items-center justify-center rounded-full text-fc-xs font-bold {k.chip}"
			aria-hidden="true">{k.mark}</span
		>
		<div class="md-flow min-w-0 leading-relaxed text-fc-fg">{@html html}</div>
	</li>
{/if}
