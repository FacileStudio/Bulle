<script lang="ts">
	import '../app.css';
	import 'blobatar/motion.css';
	import 'blobatar/gaze.css';
	import '$lib/icons';
	import { onMount } from 'svelte';
	import { Toaster } from '@facile/muse';
	import { registerDeepLinks } from '$lib/native';
	import { isNative } from '$lib/platform';

	let { children } = $props();

	onMount(() => {
		if (isNative()) registerDeepLinks();
	});
</script>

<Toaster class="pb-28 md:pb-6" />
{@render children()}
