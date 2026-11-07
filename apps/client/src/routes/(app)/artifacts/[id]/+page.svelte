<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button, ConfirmModal, PageHeader, Stack, icons, toast } from '@facile/muse';
	import { artifactError, deleteArtifact, type ArtifactDetail } from '$lib/artifacts';
	import ArtifactView from '$lib/components/ArtifactView.svelte';
	import { spaces } from '$lib/spaces.svelte';

	let artifact = $state<ArtifactDetail | null>(null);
	let confirmOpen = $state(false);

	const id = $derived(page.params.id ?? '');
	const updated = $derived(
		artifact ? new Date(artifact.updated).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' }) : ''
	);

	async function remove() {
		try {
			await deleteArtifact(id, spaces.currentId);
			confirmOpen = false;
			toast.success('Artifact deleted');
			await goto('/artifacts');
		} catch (err) {
			confirmOpen = false;
			toast.danger(artifactError(err).text);
		}
	}
</script>

<svelte:head>
	<title>{artifact?.title ?? 'Artifact'} - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader
		title={artifact?.title ?? 'Artifact'}
		description={updated ? `Updated ${updated}` : undefined}
		back={{ href: '/artifacts', label: 'Artifacts' }}
	>
		{#snippet actions()}
			{#if artifact}
				<Button variant="ghost-danger" icon={icons.remove} class="h-11" onclick={() => (confirmOpen = true)}>
					Delete
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
	<article class="min-w-0">
		{#key id}
			<ArtifactView {id} spaceId={spaces.currentId} onload={(a) => (artifact = a)} />
		{/key}
	</article>
</Stack>

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title="Delete this artifact?"
	description="It is removed from Mycelium as well. This cannot be undone."
	confirmLabel="Delete"
	onConfirm={remove}
	onCancel={() => (confirmOpen = false)}
/>
