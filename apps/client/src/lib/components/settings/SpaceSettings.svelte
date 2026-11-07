<script lang="ts">
	import { Badge, Button, Card, ConfirmModal, Input, Section, icons, toast } from '@facile/muse';
	import CreateSpaceDialog from '$lib/components/CreateSpaceDialog.svelte';
	import { spaces, type Space } from '$lib/spaces.svelte';
	import { t } from '$lib/i18n/index.svelte';

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
			toast.success(t('spaces.renamed'));
			renaming = null;
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('spaces.renameFailed'));
		}
	}

	/** Deletes the space the user confirmed. */
	async function confirmDelete() {
		const space = pendingDelete;
		confirmOpen = false;
		if (!space) return;
		try {
			await spaces.remove(space.id);
			toast.success(t('spaces.deleted', { name: space.name }));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('spaces.deleteFailed'));
		} finally {
			pendingDelete = null;
		}
	}
</script>

<Section
	card
	title={t('spaces.title')}
	description={t('spaces.description')}
>
	{#snippet actions()}
		<Button icon={icons.plus} onclick={() => (createOpen = true)}>{t('spaces.new')}</Button>
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
						<Input bind:value={name} maxlength={60} required aria-label={t('spaces.nameAria')} />
						<Button type="submit" variant="primary" icon={icons.check}>{t('spaces.save')}</Button>
						<Button type="button" variant="ghost" onclick={() => (renaming = null)}>{t('common.cancel')}</Button>
					</form>
				{:else}
					<span class="min-w-0 flex-1 truncate font-medium text-fc-fg">{space.name}</span>
					{#if space.id === spaces.currentId}
						<Badge tone="success">{t('spaces.current')}</Badge>
					{/if}
					{#if shared}
						<Badge>{t('spaces.shared')}</Badge>
					{:else}
						<Button variant="ghost" icon={icons.edit} onclick={() => startRename(space)}>{t('spaces.rename')}</Button>
						<Button
							variant="ghost"
							icon={icons.remove}
							disabled={spaces.items.length <= 1}
							onclick={() => {
								pendingDelete = space;
								confirmOpen = true;
							}}
						>
							{t('spaces.delete')}
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
	title={pendingDelete ? t('spaces.deleteTitle', { name: pendingDelete.name }) : t('spaces.deleteTitleFallback')}
	description={t('spaces.deleteDescription')}
	confirmLabel={t('spaces.deleteConfirm')}
	onConfirm={confirmDelete}
	onCancel={() => {
		confirmOpen = false;
		pendingDelete = null;
	}}
/>
