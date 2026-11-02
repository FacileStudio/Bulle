<script lang="ts">
	import { Blobatar } from '@blobatar/svelte';
	import { gaze } from '@blobatar/svelte/gaze';
	import { caretGazeTarget, type CaretCoordinates } from '$lib/caret';

	const BASE = 192;

	let {
		name,
		size,
		caret = null
	}: { name: string; size: number; caret?: CaretCoordinates | null } = $props();

	let box = $state<HTMLElement | null>(null);
	const eyes = gaze({ travel: 6, target: 'pointer' });

	$effect(() => {
		const target = caret && box ? caretGazeTarget(box.getBoundingClientRect(), caret) : null;
		eyes.lookAt(target ?? 'pointer');
	});

	$effect(() => {
		function onpointermove() {
			if (!caret) eyes.lookAt('pointer');
		}
		window.addEventListener('pointermove', onpointermove, { passive: true });
		return () => window.removeEventListener('pointermove', onpointermove);
	});
</script>

<div
	bind:this={box}
	class="relative shrink-0 select-none"
	style:width="{size}px"
	style:height="{size}px"
	style:transition="width 600ms var(--ease-fc), height 600ms var(--ease-fc)"
>
	<div
		class="absolute top-0 left-0 origin-top-left"
		style:transform="scale({size / BASE})"
		style:transition="transform 600ms var(--ease-fc)"
	>
		<Blobatar {@attach eyes} {name} size={BASE} animate="always" class="drop-shadow-sm" />
	</div>
</div>
