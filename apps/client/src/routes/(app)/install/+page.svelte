<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { PageHeader } from '@facile/muse';
	import Downloads from '$lib/components/Downloads.svelte';
</script>

<svelte:head>
	<title>{t('downloads.install.pageTitle')}</title>
</svelte:head>

<div class="flex flex-col gap-8">
	<PageHeader
		title={t('downloads.install.title')}
		description={t('downloads.install.description')}
	/>
	<Downloads headingLevel={3} />
</div>
