<script lang="ts">
	import { renderMarkdown } from '$lib/chat/markdown';

	let { text, streaming = false }: { text: string; streaming?: boolean } = $props();

	const html = $derived(renderMarkdown(text));
</script>

<div class="reply text-fc-md text-fc-fg" aria-busy={streaming}>
	{@html html}
</div>

<style>
	.reply {
		overflow-wrap: anywhere;
	}
	.reply :global(> *:first-child) {
		margin-top: 0;
	}
	.reply :global(> *:last-child) {
		margin-bottom: 0;
	}
	.reply :global(p),
	.reply :global(ul),
	.reply :global(ol),
	.reply :global(pre),
	.reply :global(blockquote),
	.reply :global(table) {
		margin: 0.75rem 0;
	}
	.reply :global(h1),
	.reply :global(h2),
	.reply :global(h3),
	.reply :global(h4) {
		margin: 1.25rem 0 0.5rem;
		font-weight: 600;
		line-height: 1.3;
	}
	.reply :global(h1) {
		font-size: var(--text-fc-xl);
	}
	.reply :global(h2) {
		font-size: var(--text-fc-lg);
	}
	.reply :global(h3),
	.reply :global(h4) {
		font-size: var(--text-fc-md);
	}
	.reply :global(ul) {
		list-style: disc;
		padding-left: 1.25rem;
	}
	.reply :global(ol) {
		list-style: decimal;
		padding-left: 1.25rem;
	}
	.reply :global(li + li) {
		margin-top: 0.25rem;
	}
	.reply :global(a) {
		text-decoration: underline;
		text-underline-offset: 2px;
	}
	.reply :global(code) {
		font-family: var(--font-fc-mono);
		font-size: 0.875em;
		background: var(--color-fc-surface);
		border-radius: var(--radius-fc-xs);
		padding: 0.1em 0.35em;
	}
	.reply :global(pre) {
		background: var(--color-fc-surface);
		border: 1px solid var(--color-fc-border);
		border-radius: var(--radius-fc-md);
		padding: 0.75rem 1rem;
		overflow-x: auto;
	}
	.reply :global(pre code) {
		background: none;
		padding: 0;
		font-size: var(--text-fc-sm);
	}
	.reply :global(blockquote) {
		border-left: 3px solid var(--color-fc-border);
		padding-left: 1rem;
		color: var(--color-fc-fg-muted);
	}
	.reply :global(table) {
		display: block;
		overflow-x: auto;
		border-collapse: collapse;
		font-size: var(--text-fc-sm);
	}
	.reply :global(th),
	.reply :global(td) {
		border: 1px solid var(--color-fc-border);
		padding: 0.35rem 0.6rem;
		text-align: left;
	}
	.reply :global(hr) {
		border: 0;
		border-top: 1px solid var(--color-fc-border);
		margin: 1rem 0;
	}
</style>
