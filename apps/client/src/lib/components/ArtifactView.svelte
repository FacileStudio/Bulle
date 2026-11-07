<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Alert, Button, Skeleton, icons } from '@facile/muse';
	import { artifactError, getArtifact, type ArtifactDetail } from '$lib/artifacts';
	import MyceliumConnect from './MyceliumConnect.svelte';

	let {
		id,
		spaceId,
		onload,
		onnavigate
	}: {
		id: string;
		spaceId: string | null | undefined;
		onload?: (artifact: ArtifactDetail) => void;
		onnavigate?: () => void;
	} = $props();

	let artifact = $state<ArtifactDetail | null>(null);
	let failure = $state<{ text: string; connect: boolean } | null>(null);
	let attempt = $state(0);

	const renderer = import('./markdown/MarkdownMuse.svelte');

	$effect(() => {
		void attempt;
		const target = id;
		const space = spaceId;
		let stale = false;
		artifact = null;
		failure = null;
		getArtifact(target, space)
			.then((detail) => {
				if (stale) return;
				artifact = detail;
				onload?.(detail);
			})
			.catch((err: unknown) => {
				if (!stale) failure = artifactError(err);
			});
		return () => {
			stale = true;
		};
	});
</script>

{#if failure?.connect}
	<MyceliumConnect {onnavigate} />
{:else if failure}
	<Alert tone="danger" title={t('artifacts.openError')}>
		<div class="flex flex-col items-start gap-3">
			<span>{failure.text}</span>
			<Button variant="ghost" icon={icons.refresh} class="h-11" onclick={() => (attempt += 1)}>{t('common.tryAgain')}</Button>
		</div>
	</Alert>
{:else if !artifact}
	<div class="flex flex-col gap-3" aria-busy="true" aria-label={t('artifacts.loadingOne')}>
		<Skeleton class="h-8 w-2/3" />
		<Skeleton class="h-4 w-full" />
		<Skeleton class="h-4 w-5/6" />
		<Skeleton class="h-24 w-full" />
	</div>
{:else}
	{#await renderer}
		<Skeleton class="h-24 w-full" />
	{:then mod}
		<mod.default content={artifact.content} />
	{:catch}
		<pre class="font-fc-mono text-fc-xs whitespace-pre-wrap text-fc-fg">{artifact.content}</pre>
	{/await}
{/if}
