<script lang="ts">
	import { onMount } from 'svelte';
	import { Button, Icon, Spinner, icons, toast } from '@facile/muse';
	import { backend, type PushConfig } from '$lib/backend';
	import { enable, permission, supported } from '$lib/notifications';

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
			if (result === 'granted') toast.success('Notifications are on');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not turn on notifications');
		} finally {
			busy = false;
		}
	}
</script>

{#if show}
	<div class="flex flex-wrap items-center gap-3 rounded-fc-lg bg-fc-component p-4">
		<Icon icon={icons.notification} size={20} class="text-fc-fg-muted" />
		<p class="min-w-0 flex-1 text-fc-sm text-fc-fg">Get a notification when a job is ready.</p>
		<Button size="sm" variant="primary" disabled={busy} onclick={turnOn}>
			{#if busy}<Spinner size="sm" />{/if}
			Turn on
		</Button>
	</div>
{/if}
