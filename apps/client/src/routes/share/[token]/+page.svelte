<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { page } from '$app/state';
	import { Alert, Skeleton, Stack } from '@facile/muse';
	import { ApiError } from '$lib/backend';
	import { getPublicShare, type PublicShare } from '$lib/shares';
	import { formatDate } from '$lib/i18n/format';
	import EncreArtifact from '$lib/components/EncreArtifact.svelte';

	let shared = $state<PublicShare | null>(null);
	let failure = $state<'missing' | 'unavailable' | null>(null);

	const token = $derived(page.params.token ?? '');
	const updated = $derived(shared ? formatDate(shared.updated, { dateStyle: 'medium', timeStyle: 'short' }) : '');

	$effect(() => {
		const target = token;
		let stale = false;
		shared = null;
		failure = null;
		getPublicShare(target)
			.then((found) => {
				if (!stale) shared = found;
			})
			.catch((err: unknown) => {
				if (!stale) failure = err instanceof ApiError && err.status === 404 ? 'missing' : 'unavailable';
			});
		return () => {
			stale = true;
		};
	});
</script>

<svelte:head>
	<title>{shared?.title ?? t('artifacts.one')} - Bulle</title>
	<meta name="robots" content="noindex, nofollow" />
	<meta name="referrer" content="no-referrer" />
</svelte:head>

<div class="flex min-h-dvh flex-col bg-fc-page text-fc-fg">
	<main class="mx-auto w-full max-w-3xl min-w-0 flex-1 px-4 py-8 sm:px-6 sm:py-12">
		{#if failure === 'missing'}
			<Stack gap="section">
				<h1 class="font-fc-title text-2xl font-bold">{t('artifacts.share.missing.title')}</h1>
				<p class="text-fc-fg-muted">{t('artifacts.share.missing.body')}</p>
			</Stack>
		{:else if failure}
			<Alert tone="warning" title={t('artifacts.share.unavailable.title')}>{t('artifacts.share.unavailable.body')}</Alert>
		{:else if !shared}
			<div class="flex flex-col gap-3" aria-busy="true" aria-label={t('artifacts.loadingOne')}>
				<Skeleton class="h-8 w-2/3" />
				<Skeleton class="h-4 w-full" />
				<Skeleton class="h-24 w-full" />
			</div>
		{:else}
			<Stack gap="section">
				<header class="flex flex-col gap-1">
					<h1 class="font-fc-title text-2xl font-bold break-words sm:text-3xl">{shared.title}</h1>
					<p class="text-fc-xs text-fc-fg-muted">{t('artifacts.updated', { date: updated })}</p>
				</header>
				<article class="min-w-0">
					<EncreArtifact content={shared.content} />
				</article>
			</Stack>
		{/if}
	</main>
	<footer class="px-4 py-6 text-center text-fc-xs text-fc-fg-muted">
		<a href="/" class="underline-offset-4 hover:underline">{t('artifacts.share.footer')}</a>
	</footer>
</div>
