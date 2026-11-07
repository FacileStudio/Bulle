<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Badge, Section, Spinner } from '@facile/muse';
	import { getMatrix, statusTone, type MatrixStatus } from '$lib/matrix';
	import { spaces } from '$lib/spaces.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let agent = $state<MatrixStatus | null>(null);
	let loadError = $state<string | null>(null);

	onMount(async () => {
		try {
			agent = await getMatrix();
		} catch (err) {
			loadError = err instanceof Error ? err.message : t('matrix.loadFailed');
		}
	});

	function spaceName(id: string): string {
		return spaces.items.find((s) => s.id === id)?.name ?? t('matrix.defaultSpace');
	}
</script>

<Section
	card
	id="matrix"
	title={t('matrix.title')}
	description={t('matrix.description')}
>
	{#if loadError}
		<Alert tone="danger">{loadError}</Alert>
	{:else if !agent}
		<div class="flex justify-center py-6"><Spinner /></div>
	{:else if !agent.available}
		<Alert tone="info">{t('matrix.unavailable')}</Alert>
	{:else}
		<div class="flex flex-col gap-4">
			<div class="flex flex-wrap items-center gap-3">
				<span class="font-mono text-fc-sm text-fc-fg">{agent.user_id || t('matrix.notCreated')}</span>
				<Badge tone={statusTone(agent.status)}>{agent.status}</Badge>
			</div>
			{#if agent.rooms.length === 0}
				<span class="text-fc-sm text-fc-fg-muted">{t('matrix.noChannels')}</span>
			{:else}
				<ul class="flex flex-col gap-2">
					{#each agent.rooms as room (room.room_id)}
						<li class="flex flex-wrap items-center justify-between gap-2 text-fc-sm">
							<span class="text-fc-fg">{room.name}</span>
							<span class="text-fc-fg-muted">{spaceName(room.space_id)}</span>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</Section>
