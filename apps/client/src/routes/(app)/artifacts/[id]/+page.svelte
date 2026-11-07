<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button, ConfirmModal, PageHeader, Stack, icons, toast } from '@facile/muse';
	import { artifactError, deleteArtifact, type ArtifactDetail } from '$lib/artifacts';
	import ArtifactView from '$lib/components/ArtifactView.svelte';
	import { spaces } from '$lib/spaces.svelte';
	import { formatDate } from '$lib/i18n/format';

	let artifact = $state<ArtifactDetail | null>(null);
	let confirmOpen = $state(false);

	const id = $derived(page.params.id ?? '');
	const updated = $derived(
		artifact ? formatDate(artifact.updated, { dateStyle: 'medium', timeStyle: 'short' }) : ''
	);

	async function remove() {
		try {
			await deleteArtifact(id, spaces.currentId);
			confirmOpen = false;
			toast.success(t('artifacts.deleted'));
			await goto('/artifacts');
		} catch (err) {
			confirmOpen = false;
			toast.danger(artifactError(err).text);
		}
	}
</script>

<svelte:head>
	<title>{artifact?.title ?? t('artifacts.one')} - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader
		title={artifact?.title ?? t('artifacts.one')}
		description={updated ? t('artifacts.updated', { date: updated }) : undefined}
		back={{ href: '/artifacts', label: t('artifacts.title') }}
	>
		{#snippet actions()}
			{#if artifact}
				<Button variant="ghost-danger" icon={icons.remove} class="h-11" onclick={() => (confirmOpen = true)}>
					{t('common.delete')}
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
	title={t('artifacts.confirm.title')}
	description={t('artifacts.confirm.description')}
	confirmLabel={t('common.delete')}
	onConfirm={remove}
	onCancel={() => (confirmOpen = false)}
/>
