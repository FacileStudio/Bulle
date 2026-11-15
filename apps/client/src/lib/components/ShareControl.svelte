<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Button, ConfirmModal, Input, Skeleton, icons, toast } from '@facile/muse';
	import { artifactError, type ArtifactWhere } from '$lib/artifacts';
	import { createShare, getShare, revokeShare, type Share } from '$lib/shares';

	let { id, where }: { id: string; where: ArtifactWhere } = $props();

	let share = $state<Share | null>(null);
	let loading = $state(true);
	let busy = $state(false);
	let copied = $state(false);
	let confirmOpen = $state(false);

	$effect(() => {
		const target = id;
		const place = { myceliumSpace: where.myceliumSpace, spaceId: where.spaceId };
		let stale = false;
		loading = true;
		getShare(target, place)
			.then((found) => {
				if (!stale) share = found;
			})
			.catch(() => {
				if (!stale) share = null;
			})
			.finally(() => {
				if (!stale) loading = false;
			});
		return () => {
			stale = true;
		};
	});

	async function create() {
		busy = true;
		try {
			share = await createShare(id, where);
		} catch (err) {
			toast.danger(artifactError(err).text, { title: t('artifacts.share.error') });
		} finally {
			busy = false;
		}
	}

	async function copy() {
		if (!share) return;
		try {
			await navigator.clipboard.writeText(share.url);
			copied = true;
			toast.success(t('artifacts.share.copied'));
			setTimeout(() => (copied = false), 2000);
		} catch {
			toast.danger(t('artifacts.share.copyError'));
		}
	}

	async function revoke() {
		busy = true;
		try {
			await revokeShare(id, where);
			share = null;
			toast.success(t('artifacts.share.revoked'));
		} catch (err) {
			toast.danger(artifactError(err).text, { title: t('artifacts.share.error') });
		} finally {
			busy = false;
			confirmOpen = false;
		}
	}
</script>

<section class="flex flex-col gap-3 rounded-fc-lg bg-fc-component p-4" aria-label={t('artifacts.share.title')}>
	{#if loading}
		<Skeleton class="h-11 w-full" />
	{:else if share}
		<div class="flex flex-col gap-2 sm:flex-row">
			<Input readonly value={share.url} aria-label={t('artifacts.share.title')} class="min-w-0 flex-1" onfocus={(e) => e.currentTarget.select()} />
			<div class="flex gap-2">
				<Button variant="secondary" icon={copied ? icons.check : icons.copy} class="h-11" onclick={copy}>
					{copied ? t('common.copied') : t('common.copy')}
				</Button>
				<Button variant="ghost-danger" icon={icons.remove} class="h-11" disabled={busy} onclick={() => (confirmOpen = true)}>
					{t('artifacts.share.revoke')}
				</Button>
			</div>
		</div>
		<p class="text-fc-xs text-fc-fg-muted">{t('artifacts.share.hint')}</p>
	{:else}
		<div class="flex flex-col items-start gap-2">
			<Button variant="secondary" icon={icons.globe} class="h-11" disabled={busy} onclick={create}>
				{t('artifacts.share.create')}
			</Button>
			<p class="text-fc-xs text-fc-fg-muted">{t('artifacts.share.hint')}</p>
		</div>
	{/if}
</section>

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title={t('artifacts.share.confirm.title')}
	description={t('artifacts.share.confirm.description')}
	confirmLabel={t('artifacts.share.revoke')}
	onConfirm={revoke}
	onCancel={() => (confirmOpen = false)}
/>
