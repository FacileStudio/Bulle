<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Alert, Button, ConfirmModal, Modal, Skeleton, Textarea, icons, toast } from '@facile/muse';
	import { deleteMemoryPage, getMemoryPage, memoryError, pageLabel, saveMemoryPage } from '$lib/memory';

	let {
		open = $bindable(false),
		path,
		spaceId,
		ondelete
	}: {
		open?: boolean;
		path: string;
		spaceId: string | null | undefined;
		ondelete?: (path: string) => void;
	} = $props();

	let content = $state<string | null>(null);
	let draft = $state('');
	let editing = $state(false);
	let saving = $state(false);
	let failure = $state<string | null>(null);
	let confirmOpen = $state(false);

	const renderer = import('../markdown/MarkdownMuse.svelte');
	const label = $derived(pageLabel(path));

	$effect(() => {
		const target = path;
		const space = spaceId;
		if (!open) return;
		let stale = false;
		content = null;
		editing = false;
		failure = null;
		getMemoryPage(target, space)
			.then((page) => {
				if (!stale) content = page.content;
			})
			.catch((err: unknown) => {
				if (!stale) failure = memoryError(err, t('memory.error.load')).text;
			});
		return () => {
			stale = true;
		};
	});

	function edit() {
		draft = content ?? '';
		editing = true;
	}

	async function save() {
		saving = true;
		try {
			const saved = await saveMemoryPage(path, draft, spaceId);
			content = saved.content;
			editing = false;
			toast.success(t('memory.saved'));
		} catch (err) {
			toast.danger(memoryError(err, t('memory.error.save')).text);
		} finally {
			saving = false;
		}
	}

	async function remove() {
		try {
			await deleteMemoryPage(path, spaceId);
			confirmOpen = false;
			open = false;
			ondelete?.(path);
			toast.success(t('memory.deleted'));
		} catch (err) {
			confirmOpen = false;
			toast.danger(memoryError(err, t('memory.error.delete')).text);
		}
	}
</script>

<Modal bind:open title={label} size="lg">
	<div class="flex flex-col gap-4">
		<p class="truncate font-fc-mono text-fc-xs text-fc-fg-muted">{path}</p>
		{#if failure}
			<Alert tone="danger">{failure}</Alert>
		{:else if content === null}
			<div class="flex flex-col gap-3" aria-busy="true" aria-label={t('memory.loading')}>
				<Skeleton class="h-6 w-2/3" />
				<Skeleton class="h-4 w-full" />
				<Skeleton class="h-4 w-5/6" />
			</div>
		{:else if editing}
			<Textarea
				bind:value={draft}
				rows={14}
				class="border-0 bg-fc-surface font-fc-mono text-fc-sm"
				aria-label={t('memory.contentLabel')}
			/>
		{:else}
			<div class="max-h-[60vh] overflow-y-auto rounded-fc-md bg-fc-surface p-4">
				{#await renderer}
					<Skeleton class="h-24 w-full" />
				{:then mod}
					<mod.default {content} />
				{:catch}
					<pre class="font-fc-mono text-fc-xs whitespace-pre-wrap text-fc-fg">{content}</pre>
				{/await}
			</div>
		{/if}

		<div class="flex flex-wrap items-center justify-between gap-2">
			<Button variant="ghost-danger" icon={icons.remove} class="h-11" disabled={content === null && !failure} onclick={() => (confirmOpen = true)}>
				{t('common.delete')}
			</Button>
			<div class="flex gap-2">
				{#if editing}
					<Button variant="ghost" class="h-11" onclick={() => (editing = false)}>{t('common.cancel')}</Button>
					<Button variant="primary" class="h-11" disabled={saving || !draft.trim()} onclick={save}>{t('common.save')}</Button>
				{:else}
					<Button variant="ghost" class="h-11" onclick={() => (open = false)}>{t('common.close')}</Button>
					<Button variant="primary" icon={icons.edit} class="h-11" disabled={content === null} onclick={edit}>
						{t('memory.edit')}
					</Button>
				{/if}
			</div>
		</div>
	</div>
</Modal>

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title={t('memory.confirm.title', { page: label })}
	description={t('memory.confirm.description')}
	confirmLabel={t('common.delete')}
	onConfirm={remove}
	onCancel={() => (confirmOpen = false)}
/>
