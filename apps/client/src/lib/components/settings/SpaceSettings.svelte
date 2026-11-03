<script lang="ts">
	import { Badge, Button, Card, ConfirmModal, Input, Section, icons, toast } from '@facile/muse';
	import CreateSpaceDialog from '$lib/components/CreateSpaceDialog.svelte';
	import { spaces, type Space } from '$lib/spaces.svelte';

	let renaming = $state<string | null>(null);
	let name = $state('');
	let createOpen = $state(false);
	let pendingDelete = $state<Space | null>(null);
	let confirmOpen = $state(false);

	/** Starts renaming one of the user's own spaces in place. */
	function startRename(space: Space) {
		renaming = space.id;
		name = space.name;
	}

	/** Saves the new name, the only thing a Bulle-only space carries. */
	async function saveName(space: Space) {
		const next = name.trim();
		if (!next || next === space.name) {
			renaming = null;
			return;
		}
		try {
			await spaces.update(space.id, { name: next });
			toast.success('Space renamed');
			renaming = null;
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not rename the space');
		}
	}

	/** Deletes the space the user confirmed. */
	async function confirmDelete() {
		const space = pendingDelete;
		confirmOpen = false;
		if (!space) return;
		try {
			await spaces.remove(space.id);
			toast.success(`${space.name} deleted`);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not delete the space');
		} finally {
			pendingDelete = null;
		}
	}
</script>

<Section
	card
	title="Spaces"
	description="Shared spaces follow your Facile spaces automatically, and the assistant uses the matching Opus workspace and Mycelium space in each. Manage members in Registre."
>
	{#snippet actions()}
		<Button icon={icons.plus} onclick={() => (createOpen = true)}>New space</Button>
	{/snippet}
	<div class="flex flex-col gap-2">
		{#each spaces.items as space (space.id)}
			{@const shared = space.registre_space_id !== ''}
			<Card class="flex items-center gap-3">
				{#if renaming === space.id}
					<form
						class="flex flex-1 items-center gap-2"
						onsubmit={(e) => {
							e.preventDefault();
							void saveName(space);
						}}
					>
						<Input bind:value={name} maxlength={60} required aria-label="Space name" />
						<Button type="submit" variant="primary" icon={icons.check}>Save</Button>
						<Button type="button" variant="ghost" onclick={() => (renaming = null)}>Cancel</Button>
					</form>
				{:else}
					<span class="min-w-0 flex-1 truncate font-medium text-fc-fg">{space.name}</span>
					{#if space.id === spaces.currentId}
						<Badge tone="success">Current</Badge>
					{/if}
					{#if shared}
						<Badge>Shared</Badge>
					{:else}
						<Button variant="ghost" icon={icons.edit} onclick={() => startRename(space)}>Rename</Button>
						<Button
							variant="ghost"
							icon={icons.remove}
							disabled={spaces.items.length <= 1}
							onclick={() => {
								pendingDelete = space;
								confirmOpen = true;
							}}
						>
							Delete
						</Button>
					{/if}
				{/if}
			</Card>
		{/each}
	</div>
</Section>

<CreateSpaceDialog bind:open={createOpen} />

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title={`Delete ${pendingDelete?.name ?? 'this space'}?`}
	description="Every chat and job in this space is deleted for good."
	confirmLabel="Delete space"
	onConfirm={confirmDelete}
	onCancel={() => {
		confirmOpen = false;
		pendingDelete = null;
	}}
/>
