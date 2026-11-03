<script lang="ts">
	import { untrack } from 'svelte';
	import {
		Badge,
		Button,
		Card,
		ConfirmModal,
		Field,
		Input,
		Section,
		icons,
		toast
	} from '@facile/muse';
	import CreateSpaceDialog from '$lib/components/CreateSpaceDialog.svelte';
	import { spaces, type Space } from '$lib/spaces.svelte';
	import { draftFor, inputFromDraft, type SpaceDraft } from '$lib/space-draft';

	let drafts = $state<Record<string, SpaceDraft>>({});
	let saving = $state<string | null>(null);
	let createOpen = $state(false);
	let pendingDelete = $state<Space | null>(null);
	let confirmOpen = $state(false);

	$effect(() => {
		const items = spaces.items;
		const previous = untrack(() => drafts);
		const next: Record<string, SpaceDraft> = {};
		for (const space of items) next[space.id] = previous[space.id] ?? draftFor(space);
		drafts = next;
	});

	async function save(space: Space) {
		const draft = drafts[space.id];
		if (!draft) return;
		saving = space.id;
		try {
			const updated = await spaces.update(space.id, inputFromDraft(draft));
			drafts[space.id] = draftFor(updated);
			toast.success('Space saved');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not save the space');
		} finally {
			saving = null;
		}
	}

	function askDelete(space: Space) {
		pendingDelete = space;
		confirmOpen = true;
	}

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
	description="Each space has its own chats, jobs and default accounts. The assistant uses a space's defaults when you do not name an account, calendar or workspace."
>
	{#snippet actions()}
		<Button icon={icons.plus} onclick={() => (createOpen = true)}>New space</Button>
	{/snippet}
	<div class="flex flex-col gap-4">
		{#each spaces.items as space (space.id)}
			{@const draft = drafts[space.id]}
			{#if draft}
				<Card>
					<form
						class="flex flex-col gap-4"
						onsubmit={(e) => {
							e.preventDefault();
							void save(space);
						}}
					>
						<div class="flex items-center gap-2">
							<span class="min-w-0 flex-1 truncate font-medium text-fc-fg">{space.name}</span>
							{#if space.id === spaces.currentId}
								<Badge tone="success">Current</Badge>
							{/if}
						</div>
						<Field label="Name" for="space-name-{space.id}">
							<Input id="space-name-{space.id}" bind:value={draft.name} maxlength={60} required />
						</Field>
						<div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
							<Field label="Courrier account id" for="space-courrier-{space.id}" helper="Empty uses your default account">
								<Input id="space-courrier-{space.id}" bind:value={draft.courrier_account} inputmode="numeric" />
							</Field>
							<Field label="Agenda calendar id" for="space-agenda-{space.id}" helper="Empty uses every calendar">
								<Input id="space-agenda-{space.id}" bind:value={draft.agenda_calendar} inputmode="numeric" />
							</Field>
							<Field label="Opus workspace" for="space-opus-{space.id}" helper="Id or name">
								<Input id="space-opus-{space.id}" bind:value={draft.opus_workspace} />
							</Field>
						</div>
						<div class="flex flex-wrap justify-end gap-2">
							<Button
								variant="outline"
								type="button"
								icon={icons.remove}
								disabled={spaces.items.length <= 1}
								onclick={() => askDelete(space)}
							>
								Delete
							</Button>
							<Button variant="primary" type="submit" disabled={saving === space.id || !draft.name.trim()}>
								{saving === space.id ? 'Saving...' : 'Save'}
							</Button>
						</div>
					</form>
				</Card>
			{/if}
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
