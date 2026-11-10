<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button, ConfirmModal, PageHeader, Skeleton, Stack, icons, toast } from '@facile/muse';
	import { artifactError, deleteArtifact, updateArtifact, type ArtifactDetail } from '$lib/artifacts';
	import ArtifactView from '$lib/components/ArtifactView.svelte';
	import { spaces } from '$lib/spaces.svelte';
	import { formatDate } from '$lib/i18n/format';

	let artifact = $state<ArtifactDetail | null>(null);
	let confirmOpen = $state(false);
	let editing = $state(false);
	let saving = $state(false);
	let revision = $state(0);

	const editor = import('$lib/components/ArtifactEditor.svelte');

	const id = $derived(page.params.id ?? '');
	const where = $derived({ myceliumSpace: page.url.searchParams.get('in'), spaceId: spaces.currentId });
	const updated = $derived(
		artifact ? formatDate(artifact.updated, { dateStyle: 'medium', timeStyle: 'short' }) : ''
	);

	async function remove() {
		try {
			await deleteArtifact(id, where);
			confirmOpen = false;
			toast.success(t('artifacts.deleted'));
			await goto('/artifacts');
		} catch (err) {
			confirmOpen = false;
			toast.danger(artifactError(err).text);
		}
	}

	async function save(markdown: string) {
		saving = true;
		try {
			artifact = await updateArtifact(id, { myceliumSpace: artifact?.mycelium_space ?? where.myceliumSpace, spaceId: where.spaceId }, { content: markdown });
			editing = false;
			revision += 1;
			toast.success(t('artifacts.saved'));
		} catch (err) {
			toast.danger(artifactError(err).text, { title: t('artifacts.saveError') });
		} finally {
			saving = false;
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
			{#if artifact && !editing}
				<Button variant="secondary" icon={icons.edit} class="h-11" onclick={() => (editing = true)}>
					{t('artifacts.edit')}
				</Button>
				<Button variant="ghost-danger" icon={icons.remove} class="h-11" onclick={() => (confirmOpen = true)}>
					{t('common.delete')}
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
	<article class="min-w-0">
		{#if editing && artifact}
			{#await editor}
				<Skeleton class="h-40 w-full" />
			{:then mod}
				<mod.default content={artifact.content} {saving} onsave={save} oncancel={() => (editing = false)} />
			{/await}
		{:else}
			{#key `${id}:${revision}`}
				<ArtifactView {id} {where} onload={(a) => (artifact = a)} />
			{/key}
		{/if}
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
