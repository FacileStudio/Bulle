<script lang="ts">
	import '../app.css';
	import 'blobatar/motion.css';
	import 'blobatar/gaze.css';
	import '$lib/icons';
	import { onMount } from 'svelte';
	import { Toaster } from '@facile/muse';
	import { registerDeepLinks } from '$lib/native';
	import { isApp, isDesktop } from '$lib/platform';
	import { offerUpdate } from '$lib/updater';

	let { children } = $props();

	onMount(() => {
		if (isApp()) registerDeepLinks();
		if (isDesktop()) void offerUpdate();
	});
</script>

<Toaster class="pb-28 md:pb-6" />
{@render children()}
