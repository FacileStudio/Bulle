<script lang="ts">
	import { locale, t } from '$lib/i18n/index.svelte';
	import { Alert, Button, Skeleton, icons, toast } from '@facile/muse';
	import { Editor, loadEncre, type EncreNode } from '@facile/encre';
	import { alertTitles, toEncreMarkdown } from './markdown/encreCompat';
	import { renderDiagram } from './markdown/encreHooks';
	import { editorRefusal, openArtifact, refusalOf } from './markdown/encreOpen';

	let {
		content,
		saving = false,
		onsave,
		oncancel
	}: {
		content: string;
		saving?: boolean;
		onsave: (markdown: string) => void;
		oncancel: () => void;
	} = $props();

	let editor = $state<ReturnType<typeof Editor>>();
	let doc = $state<EncreNode | null>(null);
	let failed = $state(false);
	let unreadable = $state<string | null>(null);

	$effect(() => {
		const source = content;
		loadEncre().then(
			() => {
				const opened = openArtifact(toEncreMarkdown(source, alertTitles()));
				doc = opened.doc;
				unreadable = opened.doc ? null : opened.reason;
			},
			() => (failed = true)
		);
	});

	function reveal(path: string, blockId?: string) {
		if (!editor) return;
		if (!(blockId && editor.revealBlock(blockId))) editor.revealPath(path);
	}

	function reported(error: Error) {
		const refusal = editorRefusal(error);
		if (refusal.kind === 'paste') {
			toast.warning(t('artifacts.pastedAsText', { reason: refusal.reason }), { title: t('artifacts.pastedAsTextTitle') });
		} else if (refusal.kind === 'other') {
			toast.danger(refusal.reason);
		}
	}

	async function save() {
		if (!editor) return;
		try {
			const markdown = await editor.exportMarkdown();
			const refusal = refusalOf(markdown);
			if (refusal) {
				toast.danger(t('artifacts.invalid', { reason: refusal }));
				return;
			}
			onsave(markdown);
		} catch (error) {
			const refusal = editorRefusal(error);
			if (refusal.kind === 'document') {
				toast.danger(t('artifacts.unsavable'), {
					title: t('artifacts.saveError'),
					duration: 0,
					action: { label: t('artifacts.show'), onClick: () => reveal(refusal.path, refusal.blockId) }
				});
				return;
			}
			toast.danger(t('artifacts.invalid', { reason: refusal.reason }));
		}
	}
</script>

{#if failed}
	<Alert tone="danger" title={t('artifacts.openError')}>{t('artifacts.editorFailed')}</Alert>
{:else if unreadable}
	<div class="flex flex-col gap-4">
		<Alert tone="danger" title={t('artifacts.openError')}>{t('artifacts.degraded', { reason: unreadable })}</Alert>
		<div class="flex flex-wrap justify-end gap-2">
			<Button variant="ghost" class="h-11" onclick={oncancel}>{t('common.cancel')}</Button>
		</div>
	</div>
{:else if !doc}
	<Skeleton class="h-40 w-full" />
{:else}
	<div class="flex flex-col gap-4">
		<Editor
			bind:this={editor}
			content={doc}
			locale={locale()}
			{renderDiagram}
			placeholder={t('artifacts.placeholder')}
			onerror={reported}
		/>
		<div class="flex flex-wrap justify-end gap-2">
			<Button variant="ghost" class="h-11" onclick={oncancel} disabled={saving}>{t('common.cancel')}</Button>
			<Button variant="primary" icon={icons.check} class="h-11" onclick={save} disabled={saving}>{t('common.save')}</Button>
		</div>
	</div>
{/if}
