<script lang="ts">
	import '../app.css';
	import '@facile/blob/motion.css';
	import '@facile/blob/gaze.css';
	import '$lib/icons';
	import '$lib/suite-icons';
	import { onMount } from 'svelte';
	import { Toaster, provideLocale } from '@facile/muse';
	import { registerDeepLinks } from '$lib/native';
	import WindowControls from '$lib/components/native/WindowControls.svelte';
	import { hasNativeTrafficLights, isApp, isDesktop } from '$lib/platform';
	import { updater } from '$lib/updater.svelte';
	import { bindLocale, localeSource } from '$lib/i18n/index.svelte';
	import { session } from '$lib/session.svelte';

	let { children } = $props();

	const i18n = provideLocale(() => localeSource(() => session.user?.locale));
	bindLocale(() => i18n.locale);

	const drawnControls = isDesktop() && !hasNativeTrafficLights();

	onMount(() => {
		if (!isApp()) return;
		registerDeepLinks();
		void updater.checkNow(true);
	});
</script>

{#if drawnControls}
	<WindowControls />
{/if}
<Toaster />
{@render children()}
