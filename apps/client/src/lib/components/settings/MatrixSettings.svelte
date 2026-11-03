<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Badge, Section, Spinner } from '@facile/muse';
	import { getMatrix, statusTone, type MatrixStatus } from '$lib/matrix';
	import { spaces } from '$lib/spaces.svelte';

	let agent = $state<MatrixStatus | null>(null);
	let loadError = $state<string | null>(null);

	onMount(async () => {
		try {
			agent = await getMatrix();
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Could not load the chat agent';
		}
	});

	function spaceName(id: string): string {
		return spaces.items.find((s) => s.id === id)?.name ?? 'Default space';
	}
</script>

<Section
	card
	id="matrix"
	title="Chat agent"
	description="Your assistant is a member of every channel of your Prisme spaces. Mention it in a channel and it answers in a thread, working in the matching space. There is nothing to set up."
>
	{#if loadError}
		<Alert tone="danger">{loadError}</Alert>
	{:else if !agent}
		<div class="flex justify-center py-6"><Spinner /></div>
	{:else if !agent.available}
		<Alert tone="info">Chat agents are not enabled on this Bulle server.</Alert>
	{:else}
		<div class="flex flex-col gap-4">
			<div class="flex flex-wrap items-center gap-3">
				<span class="font-mono text-fc-sm text-fc-fg">{agent.user_id || 'Not created yet'}</span>
				<Badge tone={statusTone(agent.status)}>{agent.status}</Badge>
			</div>
			{#if agent.rooms.length === 0}
				<span class="text-fc-sm text-fc-fg-muted">No channels yet. Create a space in Prisme and the agent joins it within ten minutes.</span>
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
