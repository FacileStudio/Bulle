<script lang="ts">
	import { marked, type Token, type Tokens } from 'marked';
	import AlertBlock from './AlertBlock.svelte';
	import CodeBlock from './CodeBlock.svelte';
	import ListItem from './ListItem.svelte';
	import MermaidBlock from './MermaidBlock.svelte';
	import RichFence from './RichFence.svelte';
	import { richBlock, richInline } from './MarkdownMuseRich';
	import { fenceKind, parseAlert, parseListItem, stripFrontmatter } from './MarkdownMuseUtils';

	let { content = '' }: { content: string } = $props();

	const tokens = $derived(marked.lexer(stripFrontmatter(content), { gfm: true }) as Token[]);

	const HEADINGS: Record<number, string> = {
		1: 'text-fc-2xl font-bold tracking-tight sm:text-fc-3xl',
		2: 'mt-4 text-fc-xl font-bold tracking-tight',
		3: 'mt-2 text-fc-lg font-semibold'
	};

	function align(value: string | null | undefined): string {
		return value === 'center' ? 'text-center' : value === 'right' ? 'text-right' : 'text-left';
	}

	function start(list: Tokens.List): number {
		return typeof list.start === 'number' ? list.start : 1;
	}
</script>

<div class="md flex min-w-0 flex-col gap-5 leading-relaxed text-fc-fg [overflow-wrap:anywhere]">
	{#each tokens as token, index (index)}
		{#if token.type === 'heading'}
			{@const h = token as Tokens.Heading}
			<svelte:element this={`h${Math.min(h.depth, 6)}`} class="text-fc-fg {HEADINGS[h.depth] ?? 'text-fc-md font-semibold'}">
				{@html richInline(h.text)}
			</svelte:element>
		{:else if token.type === 'paragraph'}
			<p class="text-fc-md text-fc-fg">{@html richInline((token as Tokens.Paragraph).text)}</p>
		{:else if token.type === 'blockquote'}
			{@const b = token as Tokens.Blockquote}
			{@const alert = parseAlert(b.text)}
			{#if alert.isAlert}
				<AlertBlock {alert} />
			{:else}
				<blockquote class="md-flow rounded-fc-lg bg-fc-surface p-4 text-fc-sm text-fc-fg-muted italic">
					{@html richBlock(b.text)}
				</blockquote>
			{/if}
		{:else if token.type === 'list'}
			{@const l = token as Tokens.List}
			{#if l.ordered}
				<ol class="flex flex-col gap-2 pl-1" start={start(l)}>
					{#each l.items as item, i (i)}
						<li class="flex items-start gap-3 text-fc-sm">
							<span
								class="flex size-5 shrink-0 items-center justify-center rounded-full bg-fc-surface font-fc-mono text-fc-xs font-semibold text-fc-accent"
								aria-hidden="true">{start(l) + i}</span
							>
							<div class="min-w-0 pt-0.5 text-fc-fg-muted">{@html richInline(parseListItem(item.text).clean)}</div>
						</li>
					{/each}
				</ol>
			{:else}
				<ul class="flex flex-col gap-2">
					{#each l.items as item, i (i)}
						{@const parsed = parseListItem(item.text)}
						<ListItem kind={parsed.kind} clean={parsed.clean} />
					{/each}
				</ul>
			{/if}
		{:else if token.type === 'table'}
			{@const t = token as Tokens.Table}
			<div class="overflow-x-auto rounded-fc-lg bg-fc-surface">
				<table class="w-full border-collapse text-fc-sm">
					<thead>
						<tr class="bg-fc-component">
							{#each t.header as cell, i (i)}
								<th class="p-3 font-semibold text-fc-fg {align(t.align[i])}">{@html richInline(cell.text)}</th>
							{/each}
						</tr>
					</thead>
					<tbody>
						{#each t.rows as row, r (r)}
							<tr class="even:bg-fc-component/50">
								{#each row as cell, i (i)}
									<td class="p-3 text-fc-fg-muted {align(t.align[i])}">{@html richInline(cell.text)}</td>
								{/each}
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else if token.type === 'code'}
			{@const c = token as Tokens.Code}
			{@const kind = fenceKind(c.lang, c.text)}
			{#if kind === 'mermaid'}
				<MermaidBlock code={c.text} />
			{:else if kind === 'code'}
				<CodeBlock code={c.text} lang={(c.lang ?? '').toLowerCase().trim()} />
			{:else}
				<RichFence {kind} source={c.text} />
			{/if}
		{:else if token.type === 'hr'}
			<div class="h-2" role="separator"></div>
		{:else if token.type === 'html'}
			<p class="text-fc-md whitespace-pre-wrap text-fc-fg">{token.raw}</p>
		{/if}
	{/each}
</div>

<style>
	.md :global(a) {
		color: var(--color-fc-accent);
		text-decoration: underline;
		text-underline-offset: 2px;
	}
	.md :global(:not(pre) > code) {
		border-radius: var(--radius-fc-xs);
		background: var(--color-fc-surface);
		padding: 0.1rem 0.35rem;
		font-family: var(--font-fc-mono);
		font-size: 0.9em;
	}
	.md :global(.md-flow) {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.md :global(.md-flow ul) {
		list-style: disc;
		padding-left: 1.25rem;
	}
	.md :global(.md-flow ol) {
		list-style: decimal;
		padding-left: 1.25rem;
	}
	.md :global(strong) {
		font-weight: 600;
		color: var(--color-fc-fg);
	}
</style>
