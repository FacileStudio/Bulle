<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Button, Icon, IconButton, Spinner, icons } from '@facile/muse';
	import { copyText } from './copy';
	import { renderMermaid } from './mermaid';

	let { code = '' }: { code: string } = $props();

	let svg = $state('');
	let error = $state<string | null>(null);
	let loading = $state(true);
	let showSource = $state(false);

	$effect(() => {
		const source = code.trim();
		let stale = false;
		loading = true;
		error = null;
		if (!source) {
			svg = '';
			loading = false;
			return;
		}
		renderMermaid(source)
			.then((out) => {
				if (!stale) svg = out;
			})
			.catch((err: unknown) => {
				if (!stale) error = err instanceof Error ? err.message : String(err);
			})
			.finally(() => {
				if (!stale) loading = false;
			});
		return () => {
			stale = true;
		};
	});
</script>

<div class="overflow-hidden rounded-fc-lg bg-fc-surface">
	<div class="flex items-center justify-between gap-2 pl-4 text-fc-xs text-fc-fg-muted">
		<span class="font-semibold tracking-wide uppercase">{t('markdown.diagram')}</span>
		<div class="flex items-center">
			{#if !error}
				<Button variant="ghost" class="h-11" onclick={() => (showSource = !showSource)}>
					{showSource ? t('markdown.diagram') : t('markdown.source')}
				</Button>
			{/if}
			<IconButton variant="ghost" aria-label={t('markdown.copyDiagram')} onclick={() => copyText(code)}>
				<Icon icon={icons.copy} size={16} />
			</IconButton>
		</div>
	</div>
	{#if showSource || error}
		<pre class="overflow-x-auto px-4 pb-4 font-fc-mono text-fc-xs leading-relaxed text-fc-fg"><code>{code}</code></pre>
		{#if error}
			<p class="bg-fc-danger/10 px-4 py-3 text-fc-xs text-fc-danger" role="alert">
				{t('markdown.diagramError', { error })}
			</p>
		{/if}
	{:else if loading}
		<div class="flex min-h-40 items-center justify-center gap-2 text-fc-xs text-fc-fg-muted">
			<Spinner size="sm" label={t('markdown.rendering')} />
		</div>
	{:else}
		<div class="overflow-x-auto px-4 pb-4 [&_svg]:mx-auto [&_svg]:h-auto [&_svg]:max-w-full">
			{@html svg}
		</div>
	{/if}
</div>
