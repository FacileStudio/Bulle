<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Icon } from '@facile/muse';
	import { CHAT_ICONS } from '$lib/icons';
	import { carriesFiles } from '$lib/chat/attachments';

	let {
		ondrop,
		disabled = false,
		class: className = '',
		children
	}: {
		ondrop: (files: File[]) => void;
		disabled?: boolean;
		class?: string;
		children: Snippet;
	} = $props();

	let depth = $state(0);
	const active = $derived(depth > 0 && !disabled);

	function enter(event: DragEvent) {
		if (!carriesFiles(event) || disabled) return;
		event.preventDefault();
		depth += 1;
	}

	function over(event: DragEvent) {
		if (!carriesFiles(event) || disabled) return;
		event.preventDefault();
		if (event.dataTransfer) event.dataTransfer.dropEffect = 'copy';
	}

	function leave(event: DragEvent) {
		if (!carriesFiles(event)) return;
		depth = Math.max(0, depth - 1);
	}

	function drop(event: DragEvent) {
		if (!carriesFiles(event)) return;
		event.preventDefault();
		depth = 0;
		if (!disabled) ondrop(Array.from(event.dataTransfer?.files ?? []));
	}

	function guard(event: DragEvent) {
		if (carriesFiles(event)) event.preventDefault();
	}
</script>

<svelte:window ondragover={guard} ondrop={guard} />

<div
	class="relative {className}"
	role="group"
	ondragenter={enter}
	ondragover={over}
	ondragleave={leave}
	ondrop={drop}
>
	{@render children()}
	{#if active}
		<div
			class="pointer-events-none absolute inset-2 z-40 flex flex-col items-center justify-center gap-3 rounded-fc-lg border-2 border-dashed border-fc-ring bg-fc-page/85 text-fc-fg backdrop-blur-sm"
			aria-hidden="true"
		>
			<Icon icon={CHAT_ICONS.file} size={32} />
			<p class="text-fc-md font-medium">Drop files to attach</p>
			<p class="text-fc-sm text-fc-fg-muted">Images, PDFs and text files, up to 20 MB each</p>
		</div>
	{/if}
</div>
