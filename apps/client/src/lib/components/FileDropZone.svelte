<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
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
			class="pointer-events-none absolute inset-2 z-40 flex flex-col items-center justify-center gap-3 rounded-fc-lg outline-2 -outline-offset-2 outline-dashed outline-fc-ring bg-fc-page/85 text-fc-fg backdrop-blur-sm"
			aria-hidden="true"
		>
			<Icon icon={CHAT_ICONS.file} size={32} />
			<p class="text-fc-md font-medium">{t('artifacts.attach.drop')}</p>
			<p class="text-fc-sm text-fc-fg-muted">{t('artifacts.attach.hint')}</p>
		</div>
	{/if}
</div>
