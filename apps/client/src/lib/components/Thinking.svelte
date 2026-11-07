<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Icon } from '@facile/muse';
	import { CHAT_ICONS } from '$lib/icons';

	let { text, active = false }: { text: string; active?: boolean } = $props();

	let open = $state(false);

	const lastLine = $derived(
		text
			.split('\n')
			.map((l) => l.trim())
			.filter(Boolean)
			.at(-1) ?? ''
	);
</script>

<div class="flex flex-col gap-1">
	<button
		type="button"
		class="flex min-h-11 w-full items-center gap-2 rounded-fc-md text-left text-fc-sm text-fc-fg-muted hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
		aria-expanded={open}
		onclick={() => (open = !open)}
	>
		<Icon
			icon={CHAT_ICONS.chevronRight}
			size={14}
			class={open ? 'rotate-90 transition-transform' : 'transition-transform'}
		/>
		<span class="shrink-0 font-medium">{active ? t('thinking.active') : t('thinking.done')}</span>
		{#if !open && lastLine}
			<span class="min-w-0 flex-1 truncate">{lastLine}</span>
		{/if}
	</button>
	{#if open}
		<p class="rounded-fc-md bg-fc-surface px-3 py-2 text-fc-sm whitespace-pre-wrap text-fc-fg-muted">
			{text}
		</p>
	{/if}
</div>
