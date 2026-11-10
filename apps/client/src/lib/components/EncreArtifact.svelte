<script lang="ts">
	import { locale, t } from '$lib/i18n/index.svelte';
	import { Alert, Skeleton } from '@facile/muse';
	import { Document, loadEncre } from '@facile/encre';
	import { alertTitles, toEncreMarkdown } from './markdown/encreCompat';
	import { highlight, renderDiagram } from './markdown/encreHooks';
	import { openArtifact } from './markdown/encreOpen';
	import { stripFrontmatter } from './markdown/MarkdownMuseUtils';

	let { content }: { content: string } = $props();

	const body = $derived(toEncreMarkdown(stripFrontmatter(content), alertTitles()));
	let wasm = $state<'loading' | 'ready' | 'failed'>('loading');

	$effect(() => {
		loadEncre().then(
			() => (wasm = 'ready'),
			() => (wasm = 'failed')
		);
	});

	const parsed = $derived(wasm === 'ready' ? openArtifact(body) : null);
</script>

{#snippet source()}
	<pre class="font-fc-mono text-fc-xs whitespace-pre-wrap text-fc-fg">{body}</pre>
{/snippet}

{#if wasm === 'loading'}
	<Skeleton class="h-24 w-full" />
{:else if wasm === 'failed' || !parsed || !parsed.doc}
	{#if parsed?.reason}
		<Alert tone="warning" title={t('artifacts.degradedTitle')} class="mb-4">{t('artifacts.degraded', { reason: parsed.reason })}</Alert>
	{/if}
	{@render source()}
{:else}
	{#if parsed.reason}
		<Alert tone="warning" title={t('artifacts.degradedTitle')} class="mb-4">{t('artifacts.degraded', { reason: parsed.reason })}</Alert>
	{/if}
	<svelte:boundary>
		<Document doc={parsed.doc} locale={locale()} {renderDiagram} {highlight} />
		{#snippet failed()}
			{@render source()}
		{/snippet}
	</svelte:boundary>
{/if}
