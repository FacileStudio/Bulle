<script lang="ts">
	import { Alert, Button, Field, Input, Modal } from '@facile/muse';
	import { spaces, type Space } from '$lib/spaces.svelte';

	let {
		open = $bindable(false),
		oncreated
	}: { open?: boolean; oncreated?: (space: Space) => void } = $props();

	let name = $state('');
	let submitting = $state(false);
	let error = $state<string | null>(null);

	$effect(() => {
		if (open) {
			name = '';
			error = null;
		}
	});

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		submitting = true;
		error = null;
		try {
			const created = await spaces.create({ name: name.trim() });
			open = false;
			oncreated?.(created);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not create the space';
		} finally {
			submitting = false;
		}
	}
</script>

<Modal bind:open title="New space" size="sm">
	<form onsubmit={submit} class="flex flex-col gap-4">
		<p class="text-fc-sm text-fc-fg-muted">
			A space keeps its own chats, jobs and default accounts apart from your other spaces.
		</p>
		{#if error}
			<Alert tone="danger">{error}</Alert>
		{/if}
		<Field label="Name" for="new-space-name">
			<Input id="new-space-name" bind:value={name} placeholder="e.g. Work" maxlength={60} required />
		</Field>
		<div class="flex justify-end gap-2">
			<Button variant="outline" type="button" onclick={() => (open = false)}>Cancel</Button>
			<Button variant="primary" type="submit" disabled={submitting || !name.trim()}>
				{submitting ? 'Creating...' : 'Create space'}
			</Button>
		</div>
	</form>
</Modal>
