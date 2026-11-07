<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, Section, Spinner, toast } from '@facile/muse';
	import { backend, type PushConfig } from '$lib/backend';
	import { webDelivery } from '$lib/live/delivery.svelte';
	import { enable, tabOnlyMessage, permission, supported, type Permission } from '$lib/notifications';
	import { isDesktop, isNative } from '$lib/platform';
	import { t } from '$lib/i18n/index.svelte';

	let push = $state<PushConfig | null>(null);
	let access = $state<Permission | 'loading'>('loading');
	let busy = $state(false);
	let testing = $state(false);
	const native = isNative();
	const desktop = isDesktop();
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
			if (access === 'granted' && !webDelivery.tabOnly) toast.success(t('notifications.on'));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('notifications.turnOnFailed'));
		} finally {
			busy = false;
		}
	}

	async function sendTest() {
		testing = true;
		try {
			await backend.testPush();
			toast.success(t('notifications.testSent'));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('notifications.testFailed'));
		} finally {
			testing = false;
		}
	}

	async function copySubscribeUrl() {
		if (!push) return;
		try {
			await navigator.clipboard.writeText(push.subscribe_url);
			toast.success(t('notifications.linkCopied'));
		} catch {
			toast.danger(t('notifications.copyFailed'));
		}
	}
</script>

<Section
	card
	title={t('notifications.title')}
	description={t('notifications.description')}
>
	<div class="flex flex-col gap-4">
		{#if access === 'loading'}
			<Spinner size="sm" />
		{:else if !supported(push)}
			<p class="text-fc-sm text-fc-fg-muted">
				{native ? t('notifications.notSetUpNative') : t('notifications.notSetUpBrowser')}
			</p>
		{:else if access === 'unsupported'}
			<p class="text-fc-sm text-fc-fg-muted">{t('notifications.unsupported')}</p>
		{:else if access === 'denied'}
			<Alert tone="warning">
				{desktop
					? t('notifications.blockedDesktop')
					: native
						? t('notifications.blockedNative')
						: t('notifications.blockedBrowser')}
			</Alert>
		{:else if access === 'prompt'}
			<div>
				<Button variant="primary" disabled={busy} onclick={turnOn}>
					{#if busy}<Spinner size="sm" />{/if}
					{t('notifications.turnOn')}
				</Button>
			</div>
		{:else if webDelivery.tabOnly}
			<Alert tone="info">{tabOnlyMessage()}</Alert>
		{:else}
			<p class="text-fc-sm text-fc-fg">{t('notifications.onForDevice')}</p>
		{/if}

		{#if push && (access === 'granted' || push.enabled)}
			<div>
				<Button variant="outline" disabled={testing} onclick={sendTest}>
					{#if testing}<Spinner size="sm" />{/if}
					{t('notifications.sendTest')}
				</Button>
			</div>
		{/if}

		{#if push?.enabled}
			<div class="flex flex-col gap-3 rounded-fc-md bg-fc-surface p-4">
				<p class="text-fc-sm text-fc-fg-muted">
					{t('notifications.ntfyHint')}
				</p>
				<dl class="grid gap-1 text-fc-sm">
					<dt class="text-fc-fg-muted">{t('notifications.server')}</dt>
					<dd class="font-mono break-all text-fc-fg">{push.server}</dd>
					<dt class="mt-2 text-fc-fg-muted">{t('notifications.topic')}</dt>
					<dd class="font-mono break-all text-fc-fg">{push.topic}</dd>
				</dl>
				<div>
					{#if native}
						<Button variant="outline" onclick={() => (window.location.href = ntfyUrl)}>{t('notifications.openNtfy')}</Button>
					{:else}
						<Button variant="outline" onclick={copySubscribeUrl}>{t('notifications.copyLink')}</Button>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</Section>
