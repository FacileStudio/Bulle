<script lang="ts">
	import { locale, t } from '$lib/i18n/index.svelte';
	import { Alert, Button, Skeleton, icons, toast } from '@facile/muse';
	import { Editor, loadEncre, type EncreNode } from '@facile/encre';
	import { alertTitles, toEncreMarkdown } from './markdown/encreCompat';
	import { renderDiagram } from './markdown/encreHooks';
	import { openArtifact } from './markdown/encreOpen';

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

	async function save() {
		if (!editor) return;
		try {
			onsave(await editor.exportMarkdown());
		} catch (error) {
			toast.danger(t('artifacts.invalid', { reason: error instanceof Error ? error.message : String(error) }));
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
			onerror={(error) => toast.danger(error.message)}
		/>
		<div class="flex flex-wrap justify-end gap-2">
			<Button variant="ghost" class="h-11" onclick={oncancel} disabled={saving}>{t('common.cancel')}</Button>
			<Button variant="primary" icon={icons.check} class="h-11" onclick={save} disabled={saving}>{t('common.save')}</Button>
		</div>
	</div>
{/if}
