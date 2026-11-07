<script lang="ts">
	let { x, y, visible, down }: { x: number; y: number; visible: boolean; down: boolean } = $props();
</script>

<svg
	class="cursor pointer-events-none absolute top-0 left-0 z-30 drop-shadow-sm"
	style:translate="{x}px {y}px"
	style:opacity={visible ? 1 : 0}
	style:scale={down ? 0.88 : 1}
	width="17"
	height="22"
	viewBox="0 0 13 19"
	aria-hidden="true"
>
	<path d="m6.148 18.473 1.863-1.003 1.615-.839-2.568-4.816h4.332l-11.379-11.408v16.015l3.316-3.221z" fill="#fff" />
	<path d="m6.431 17 1.765-.941-2.775-5.202h3.604l-8.025-8.043v11.188l2.53-2.442z" fill="#000" />
</svg>

<style>
	.cursor {
		transform-origin: 0 0;
		transition:
			translate 0.9s cubic-bezier(0.45, 0, 0.1, 1),
			opacity 0.3s ease,
			scale 0.12s ease;
	}
</style>
