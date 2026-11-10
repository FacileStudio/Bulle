<script lang="ts">
	import { locale, t } from '$lib/i18n/index.svelte';
	import { Alert, Skeleton } from '@facile/muse';
	import { Document, loadEncre, parseMarkdown, validateDocument, type EncreNode } from '@facile/encre';
	import { alertTitles, toEncreMarkdown } from './markdown/encreCompat';
	import { highlight, renderDiagram } from './markdown/encreHooks';
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

	const parsed = $derived.by((): { doc: EncreNode | null; reason: string | null } | null => {
		if (wasm !== 'ready') return null;
		try {
			const doc = parseMarkdown(body);
			return { doc, reason: validateDocument(doc) };
		} catch (error) {
			return { doc: null, reason: error instanceof Error ? error.message : String(error) };
		}
	});
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
