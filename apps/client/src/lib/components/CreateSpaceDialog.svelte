<script lang="ts">
	import { Alert, Button, Field, Input, Modal } from '@facile/muse';
	import { spaces, type Space } from '$lib/spaces.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let {
		open = $bindable(false),
		oncreated
	}: { open?: boolean; oncreated?: (space: Space) => void } = $props();

	let name = $state('');
	let submitting = $state(false);
	let error = $state<string | null>(null);

	$effect(() => {
		if (open) {
			name = '';
			error = null;
		}
	});

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		submitting = true;
		error = null;
		try {
			const created = await spaces.create({ name: name.trim() });
			open = false;
			oncreated?.(created);
		} catch (err) {
			error = err instanceof Error ? err.message : t('spaces.create.failed');
		} finally {
			submitting = false;
		}
	}
</script>

<Modal bind:open title={t('spaces.create.title')} size="sm">
	<form onsubmit={submit} class="flex flex-col gap-4">
		<p class="text-fc-sm text-fc-fg-muted">
			{t('spaces.create.intro')}
		</p>
		{#if error}
			<Alert tone="danger">{error}</Alert>
		{/if}
		<Field label={t('spaces.create.name')} for="new-space-name">
			<Input id="new-space-name" bind:value={name} placeholder={t('spaces.create.placeholder')} maxlength={60} required />
		</Field>
		<div class="flex justify-end gap-2">
			<Button variant="secondary" type="button" onclick={() => (open = false)}>{t('common.cancel')}</Button>
			<Button variant="primary" type="submit" disabled={submitting || !name.trim()}>
				{submitting ? t('spaces.create.creating') : t('spaces.create.submit')}
			</Button>
		</div>
	</form>
</Modal>
