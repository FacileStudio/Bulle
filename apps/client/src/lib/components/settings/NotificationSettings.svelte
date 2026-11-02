<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, Section, Spinner, toast } from '@facile/muse';
	import { backend, type PushConfig } from '$lib/backend';
	import { enable, permission, supported, type Permission } from '$lib/notifications';
	import { isNative } from '$lib/platform';

	let push = $state<PushConfig | null>(null);
	let access = $state<Permission | 'loading'>('loading');
	let busy = $state(false);
	let testing = $state(false);
	const native = isNative();
	const ntfyUrl = $derived(
		push?.enabled ? `ntfy://${new URL(push.server).host}/${push.topic}` : ''
	);

	onMount(async () => {
		try {
			push = await backend.getPush();
		} catch {
			push = null;
		}
		access = await permission().catch(() => 'unsupported' as const);
	});

	async function turnOn() {
		busy = true;
		try {
			access = await enable(push);
			if (access === 'granted') toast.success('Notifications are on');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not turn on notifications');
		} finally {
			busy = false;
		}
	}

	async function sendTest() {
		testing = true;
		try {
			await backend.testPush();
			toast.success('Test notification sent');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not send the test');
		} finally {
			testing = false;
		}
	}

	async function copySubscribeUrl() {
		if (!push) return;
		try {
			await navigator.clipboard.writeText(push.subscribe_url);
			toast.success('Link copied');
		} catch {
			toast.danger('Could not copy the link');
		}
	}
</script>

<Section
	card
	title="Notifications"
	description="Get a notification when a job finishes or your assistant needs an approval."
>
	<div class="flex flex-col gap-4">
		{#if access === 'loading'}
			<Spinner size="sm" />
		{:else if !supported(push)}
			<p class="text-fc-sm text-fc-fg-muted">
				{native ? 'Native notifications' : 'Browser notifications'} are not set up on this server yet.
			</p>
		{:else if access === 'unsupported'}
			<p class="text-fc-sm text-fc-fg-muted">This browser cannot show notifications.</p>
		{:else if access === 'denied'}
			<Alert tone="warning">
				{native
					? 'Notifications are blocked. Turn them on in Settings > Apps > Bulle > Notifications.'
					: 'Notifications are blocked for this site. Allow them in your browser settings.'}
			</Alert>
		{:else if access === 'prompt'}
			<div>
				<Button variant="primary" disabled={busy} onclick={turnOn}>
					{#if busy}<Spinner size="sm" />{/if}
					Turn on notifications
				</Button>
			</div>
		{:else}
			<p class="text-fc-sm text-fc-fg">Notifications are on for this device.</p>
		{/if}

		{#if push && (access === 'granted' || push.enabled)}
			<div>
				<Button variant="outline" disabled={testing} onclick={sendTest}>
					{#if testing}<Spinner size="sm" />{/if}
					Send test
				</Button>
			</div>
		{/if}

		{#if push?.enabled}
			<div class="flex flex-col gap-3 rounded-fc-md bg-fc-surface p-4">
				<p class="text-fc-sm text-fc-fg-muted">
					You can also follow your private ntfy topic from the ntfy app.
				</p>
				<dl class="grid gap-1 text-fc-sm">
					<dt class="text-fc-fg-muted">Server</dt>
					<dd class="font-mono break-all text-fc-fg">{push.server}</dd>
					<dt class="mt-2 text-fc-fg-muted">Topic</dt>
					<dd class="font-mono break-all text-fc-fg">{push.topic}</dd>
				</dl>
				<div>
					{#if native}
						<Button variant="outline" onclick={() => (window.location.href = ntfyUrl)}>Open in ntfy</Button>
					{:else}
						<Button variant="outline" onclick={copySubscribeUrl}>Copy subscribe link</Button>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</Section>
