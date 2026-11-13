<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import Downloads from '$lib/components/Downloads.svelte';
	import PublicShell from '$lib/components/PublicShell.svelte';
</script>

<svelte:head>
	<title>{t('downloads.pageTitle')}</title>
	<meta
		name="description"
		content={t('downloads.metaDescription')}
	/>
</svelte:head>

<PublicShell>
	<section class="mx-auto max-w-5xl px-6 pt-16 pb-10 text-center sm:pt-20">
		<h1 class="font-fc-title text-3xl font-bold tracking-tight text-fc-fg sm:text-5xl">
			{t('downloads.hero.title')}
		</h1>
		<p class="mx-auto mt-4 max-w-xl text-fc-sm text-fc-fg-muted sm:text-fc-base">
			{t('downloads.hero.description')}
		</p>
	</section>
	<div class="mx-auto max-w-5xl px-6 pb-20">
		<Downloads />
	</div>
</PublicShell>
