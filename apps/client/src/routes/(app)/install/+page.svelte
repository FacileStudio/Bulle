<script lang="ts">
	import { PageHeader } from '@facile/muse';
	import DownloadOptions from '$lib/components/DownloadOptions.svelte';
</script>

<svelte:head>
	<title>Get the apps · Bulle</title>
</svelte:head>

<div class="flex flex-col gap-8">
	<PageHeader
		title="Get the apps"
		description="Use Bulle on your desktop, your phone and your terminal. You stay signed in to the same account."
	/>
	<DownloadOptions />
</div>
