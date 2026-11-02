<script lang="ts">
	import '../app.css';
	import 'blobatar/motion.css';
	import 'blobatar/gaze.css';
	import '$lib/icons';
	import { Toaster } from '@facile/muse';

	let { children } = $props();
</script>

<Toaster class="pb-28 md:pb-6" />
{@render children()}
