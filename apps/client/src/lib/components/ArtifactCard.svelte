<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Icon, icons } from '@facile/muse';
	import { artifactPanel, openArtifact } from '$lib/artifact-panel.svelte';
	import type { ArtifactRef } from '$lib/artifacts';
	import { CHAT_ICONS } from '$lib/icons';

	let { artifact }: { artifact: ArtifactRef } = $props();

	const open = $derived(artifactPanel.ref?.id === artifact.id);
</script>

<button
	type="button"
	class="flex min-h-14 w-full max-w-md items-center gap-3 rounded-fc-lg bg-fc-surface px-4 py-2 text-left transition-colors hover:bg-fc-component focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring {open
		? 'bg-fc-component'
		: ''}"
	aria-haspopup="dialog"
	aria-expanded={open}
	onclick={(event) => openArtifact(artifact, event.currentTarget)}
>
	<span class="flex size-9 shrink-0 items-center justify-center rounded-fc-md bg-fc-accent/15 text-fc-accent">
		<Icon icon={icons.layers} size={18} />
	</span>
	<span class="flex min-w-0 flex-1 flex-col">
		<span class="truncate text-fc-sm font-medium text-fc-fg">{artifact.title}</span>
		<span class="text-fc-xs text-fc-fg-muted">{t('artifacts.one')}</span>
	</span>
	<Icon icon={CHAT_ICONS.chevronRight} size={16} class="shrink-0 text-fc-fg-muted" />
</button>
