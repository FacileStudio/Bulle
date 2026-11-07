<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { PageHeader } from '@facile/muse';
	import DownloadOptions from '$lib/components/DownloadOptions.svelte';
</script>

<svelte:head>
	<title>{t('downloads.install.pageTitle')}</title>
</svelte:head>

<div class="flex flex-col gap-8">
	<PageHeader
		title={t('downloads.install.title')}
		description={t('downloads.install.description')}
	/>
	<DownloadOptions />
</div>
