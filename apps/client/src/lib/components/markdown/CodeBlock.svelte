<script lang="ts">
	import { Icon, IconButton, icons } from '@facile/muse';
	import hljs from 'highlight.js/lib/common';
	import { copyText } from './copy';
	import { parseDiff } from './MarkdownMuseUtils';

	let { code, lang = '' }: { code: string; lang?: string } = $props();

	const isDiff = $derived(lang === 'diff');
	const diffLines = $derived(isDiff ? parseDiff(code) : []);
	const highlighted = $derived(
		!isDiff && lang && hljs.getLanguage(lang)
			? hljs.highlight(code, { language: lang, ignoreIllegals: true }).value
			: ''
	);

	const DIFF_TONES = {
		add: 'bg-fc-success/15 text-fc-success',
		del: 'bg-fc-danger/15 text-fc-danger',
		hunk: 'bg-fc-info/15 text-fc-info italic',
		ctx: 'text-fc-fg-muted'
	} as const;
</script>

<div class="overflow-hidden rounded-fc-lg bg-fc-surface">
	<div class="flex items-center justify-between gap-2 pl-4 text-fc-xs font-semibold tracking-wide text-fc-fg-muted uppercase">
		<span class="truncate">{lang || 'code'}</span>
		<IconButton variant="ghost" aria-label="Copy the code" onclick={() => copyText(code)}>
			<Icon icon={icons.copy} size={16} />
		</IconButton>
	</div>
	{#if isDiff}
		<pre class="overflow-x-auto px-2 pb-3 font-fc-mono text-fc-xs leading-relaxed"><code
				>{#each diffLines as dl, i (i)}<span class="block px-2 {DIFF_TONES[dl.type]}">{dl.line || ' '}</span
					>{/each}</code
			></pre>
	{:else if highlighted}
		<pre class="md-code overflow-x-auto px-4 pb-4 font-fc-mono text-fc-xs leading-relaxed text-fc-fg"><code
				>{@html highlighted}</code
			></pre>
	{:else}
		<pre class="overflow-x-auto px-4 pb-4 font-fc-mono text-fc-xs leading-relaxed text-fc-fg"><code>{code}</code></pre>
	{/if}
</div>

<style>
	.md-code :global(.hljs-keyword),
	.md-code :global(.hljs-selector-tag),
	.md-code :global(.hljs-built_in),
	.md-code :global(.hljs-type) {
		color: var(--color-fc-accent);
	}
	.md-code :global(.hljs-string),
	.md-code :global(.hljs-attr),
	.md-code :global(.hljs-regexp) {
		color: var(--color-fc-success);
	}
	.md-code :global(.hljs-number),
	.md-code :global(.hljs-literal),
	.md-code :global(.hljs-symbol) {
		color: var(--color-fc-warning);
	}
	.md-code :global(.hljs-title),
	.md-code :global(.hljs-function),
	.md-code :global(.hljs-section) {
		color: var(--color-fc-info);
	}
	.md-code :global(.hljs-comment),
	.md-code :global(.hljs-quote),
	.md-code :global(.hljs-meta) {
		color: var(--color-fc-fg-muted);
		font-style: italic;
	}
	.md-code :global(.hljs-deletion) {
		color: var(--color-fc-danger);
	}
	.md-code :global(.hljs-addition) {
		color: var(--color-fc-success);
	}
</style>
