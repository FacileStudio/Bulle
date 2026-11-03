<script lang="ts">
	import '../app.css';
	import '@facile/blob/motion.css';
	import '@facile/blob/gaze.css';
	import '$lib/icons';
	import { onMount } from 'svelte';
	import { Toaster } from '@facile/muse';
	import { registerDeepLinks } from '$lib/native';
	import WindowControls from '$lib/components/native/WindowControls.svelte';
	import { hasNativeTrafficLights, isApp, isDesktop } from '$lib/platform';
	import { updater } from '$lib/updater.svelte';

	let { children } = $props();

	const drawnControls = isDesktop() && !hasNativeTrafficLights();

	onMount(() => {
		if (isApp()) registerDeepLinks();
		if (isDesktop()) void updater.checkNow(true);
	});
</script>

{#if drawnControls}
	<WindowControls />
{/if}
<Toaster class="pb-28 md:pb-6" />
{@render children()}
