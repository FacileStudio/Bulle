<script lang="ts">
	import DownloadOptions from '$lib/components/DownloadOptions.svelte';
	import PublicShell from '$lib/components/PublicShell.svelte';
</script>

<svelte:head>
	<title>Downloads · Bulle</title>
	<meta
		name="description"
		content="Get Bulle for macOS, Windows, Linux, Android, or your terminal."
	/>
</svelte:head>

<PublicShell>
	<section class="mx-auto max-w-5xl px-6 pt-16 pb-10 text-center sm:pt-20">
		<h1 class="font-fc-title text-3xl font-semibold tracking-tight text-fc-fg sm:text-5xl">
			Bulle, everywhere you work
		</h1>
		<p class="mx-auto mt-4 max-w-xl text-fc-sm text-fc-fg-muted sm:text-fc-base">
			One Facile account, the same sessions and jobs on your desktop, your phone and your terminal.
		</p>
	</section>
	<div class="mx-auto max-w-5xl px-6 pb-20">
		<DownloadOptions />
	</div>
</PublicShell>
