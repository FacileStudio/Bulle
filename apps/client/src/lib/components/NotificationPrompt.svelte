<script lang="ts">
	import { onMount } from 'svelte';
	import { t } from '$lib/i18n/index.svelte';
	import { Button, Icon, Spinner, icons, toast } from '@facile/muse';
	import { backend, type PushConfig } from '$lib/backend';
	import { webDelivery } from '$lib/live/delivery.svelte';
	import { enable, tabOnlyMessage, permission, supported } from '$lib/notifications';

	let push = $state<PushConfig | null>(null);
	let show = $state(false);
	let busy = $state(false);

	onMount(async () => {
		try {
			push = await backend.getPush();
			show = supported(push) && (await permission()) === 'prompt';
		} catch {
			show = false;
		}
	});

	async function turnOn() {
		busy = true;
		try {
			const result = await enable(push);
			show = result === 'prompt';
			if (result === 'granted' && webDelivery.tabOnly) toast.info(tabOnlyMessage());
			else if (result === 'granted') toast.success(t('notifications.on'));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('notifications.turnOnFailed'));
		} finally {
			busy = false;
		}
	}
</script>

{#if show}
	<div class="flex flex-wrap items-center gap-3 rounded-fc-lg bg-fc-component p-4">
		<Icon icon={icons.notification} size={20} class="text-fc-fg-muted" />
		<p class="min-w-0 flex-1 text-fc-sm text-fc-fg">{t('notifications.prompt.text')}</p>
		<Button size="sm" variant="primary" disabled={busy} onclick={turnOn}>
			{#if busy}<Spinner size="sm" />{/if}
			{t('notifications.prompt.turnOn')}
		</Button>
	</div>
{/if}
