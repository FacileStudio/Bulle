<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Alert,
		Badge,
		Button,
		ConfirmModal,
		Icon,
		Input,
		Section,
		Spinner,
		StatusDot,
		icons,
		toast
	} from '@facile/muse';
	import { relativeTime } from '$lib/chat/format';
	import { listDevices, revokeDevice, type HardwareDevice } from '$lib/hardware/api';
	import {
		bluetoothAvailable,
		connectBohsDevice,
		scanAndPairBohsDevice,
		type PairedDevice,
		type RelayState
	} from '$lib/hardware/ble';
	import { HARDWARE_ICONS } from '$lib/icons';
	import { isApp } from '$lib/platform';
	import { t } from '$lib/i18n/index.svelte';

	let devices = $state<HardwareDevice[]>([]);
	let loadError = $state('');
	let loading = $state(true);
	let pairing = $state(false);
	let progress = $state('');
	let wifiSsid = $state('');
	let wifiPsk = $state('');
	let relays = $state.raw<PairedDevice[]>([]);
	let relayStatus = $state<Record<string, string>>({});
	let connecting = $state('');
	let pendingRevoke = $state<HardwareDevice | null>(null);
	let confirmOpen = $state(false);
	const supported = bluetoothAvailable() && !isApp();
	const transports: Record<string, string> = { ble: 'Bluetooth', wifi: 'Wi-Fi', usb: 'USB' };

	/** Reloads the user's paired devices. */
	async function load() {
		try {
			devices = await listDevices();
			loadError = '';
		} catch (err) {
			loadError = err instanceof Error ? err.message : t('hardware.loadFailed');
		} finally {
			loading = false;
		}
	}

	/** Records a relay's latest status and forgets it once it stops for good. */
	function onRelay(deviceId: string, state: RelayState) {
		relayStatus = { ...relayStatus, [deviceId]: state.message };
		if (!state.live) relays = relays.filter((r) => r.device.id !== deviceId);
	}

	/** Answers whether this page relays the device right now. */
	function relaying(device: HardwareDevice): boolean {
		return relays.some((r) => r.device.id === device.id);
	}

	/** Answers the tone for a device's relay line: healthy, degraded, or stopped. */
	function relayTone(device: HardwareDevice): 'danger' | 'success' | 'warning' {
		if (!relaying(device)) return 'danger';
		return relayStatus[device.id] === t('ble.relaying') ? 'success' : 'warning';
	}

	/** Pairs a Pastille over Bluetooth and keeps relaying it while this page is open. */
	async function pair() {
		pairing = true;
		try {
			const paired = await scanAndPairBohsDevice({
				wifiSsid: wifiSsid.trim() || undefined,
				wifiPsk: wifiPsk || undefined,
				onStatus: (message) => (progress = message),
				onRelay
			});
			relays = [...relays, paired];
			toast.success(t('hardware.paired', { name: paired.device.name }));
			await load();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('hardware.pairFailed'));
		} finally {
			pairing = false;
			progress = '';
		}
	}

	/** Relays an already paired device over Bluetooth, without pairing it again. */
	async function connectDevice(device: HardwareDevice) {
		connecting = device.id;
		try {
			const relay = await connectBohsDevice(device, { onStatus: (message) => (progress = message), onRelay });
			relays = [...relays, relay];
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('hardware.connectFailed'));
		} finally {
			connecting = '';
			progress = '';
		}
	}

	/** Stops relaying a device from this page. */
	function disconnectDevice(device: HardwareDevice) {
		relays.filter((r) => r.device.id === device.id).forEach((r) => r.close());
	}

	/** Revokes the device the user confirmed. */
	async function confirmRevoke() {
		const device = pendingRevoke;
		confirmOpen = false;
		pendingRevoke = null;
		if (!device) return;
		try {
			await revokeDevice(device.id);
			relays.filter((r) => r.device.id === device.id).forEach((r) => r.close());
			relays = relays.filter((r) => r.device.id !== device.id);
			toast.success(t('hardware.revoked', { name: device.name }));
			await load();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('hardware.revokeFailed'));
		}
	}

	onMount(() => {
		void load();
		return () => relays.forEach((r) => r.close());
	});
</script>

<Section
	card
	id="devices"
	title={t('hardware.title')}
	description={t('hardware.description')}
>
	{#snippet actions()}
		<Button icon={HARDWARE_ICONS.bluetooth} disabled={!supported || pairing} onclick={pair}>
			{#if pairing}<Spinner size="sm" />{t('hardware.pairing')}{:else}{t('hardware.pair')}{/if}
		</Button>
	{/snippet}
	<div class="flex flex-col gap-5">
		{#if !supported}
			<Alert tone="warning">
				{isApp()
					? t('hardware.unsupportedApp')
					: t('hardware.unsupportedBrowser')}
			</Alert>
		{/if}
		{#if loadError}
			<Alert tone="danger">{loadError}</Alert>
		{/if}
		{#if progress}
			<Alert tone="info">
				<span class="flex items-center gap-2">
					<Spinner size="sm" />
					{progress}
				</span>
			</Alert>
		{/if}
		<div class="flex flex-col gap-3 rounded-fc-md bg-fc-surface p-4">
			<div class="flex flex-col gap-1">
				<div class="flex items-center gap-2">
					<Icon icon={HARDWARE_ICONS.wifi} size={16} class="text-fc-fg-muted" />
					<span class="text-fc-sm font-medium text-fc-fg">{t('hardware.wifi')}</span>
				</div>
				<span class="text-fc-xs text-fc-fg-muted">
					{t('hardware.wifiHelp')}
				</span>
			</div>
			<div class="flex flex-col gap-2 sm:flex-row">
				<Input
					bind:value={wifiSsid}
					autocomplete="off"
					aria-label={t('hardware.wifiNetwork')}
					placeholder={t('hardware.wifiNetwork')}
					class="border-0"
				/>
				<Input
					type="password"
					bind:value={wifiPsk}
					autocomplete="off"
					aria-label={t('hardware.wifiPassword')}
					placeholder={t('hardware.wifiPassword')}
					class="border-0"
				/>
			</div>
		</div>
		{#if loading}
			<div class="flex justify-center py-4"><Spinner /></div>
		{:else}
			<div class="flex flex-col gap-2">
				{#each devices as device (device.id)}
					{@const relay = relayStatus[device.id]}
					<div
						class="flex flex-col gap-3 rounded-fc-md bg-fc-surface p-4 sm:flex-row sm:items-center sm:justify-between"
					>
						<div class="flex min-w-0 flex-1 items-center gap-3">
							<div
								class="flex size-9 shrink-0 items-center justify-center rounded-fc-md bg-fc-page text-fc-fg"
							>
								<Icon icon={HARDWARE_ICONS.device} size={18} />
							</div>
							<div class="flex min-w-0 flex-col gap-0.5">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{device.name}</span>
								{#if relay}
									{@const tone = relayTone(device)}
									<StatusDot {tone} pulse={tone === 'warning'} label={relay} />
								{:else}
									<span class="text-fc-xs text-fc-fg-muted">
										{device.last_seen_at
											? t('hardware.seen', { time: relativeTime(device.last_seen_at) })
											: t('hardware.neverConnected')}
									</span>
								{/if}
							</div>
						</div>
						<div class="flex flex-wrap items-center gap-2">
							{#if device.last_transport}
								<Badge>{transports[device.last_transport]}</Badge>
							{/if}
							<Button
								variant="ghost"
								size="sm"
								icon={icons.revoke}
								onclick={() => {
									pendingRevoke = device;
									confirmOpen = true;
								}}
							>
								{t('hardware.revoke')}
							</Button>
						</div>
					</div>
				{:else}
					<div
						class="flex flex-col items-center gap-2 rounded-fc-md bg-fc-surface px-4 py-8 text-center"
					>
						<div
							class="flex size-10 items-center justify-center rounded-fc-md bg-fc-page text-fc-fg-muted"
						>
							<Icon icon={HARDWARE_ICONS.device} size={20} />
						</div>
						<p class="text-fc-sm font-medium text-fc-fg">{t('hardware.empty.title')}</p>
						<p class="text-fc-xs text-fc-fg-muted">{t('hardware.empty.hint')}</p>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</Section>

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title={pendingRevoke ? t('hardware.revokeTitle', { name: pendingRevoke.name }) : t('hardware.revokeTitleFallback')}
	description={t('hardware.revokeDescription')}
	confirmLabel={t('hardware.revokeConfirm')}
	onConfirm={confirmRevoke}
	onCancel={() => {
		confirmOpen = false;
		pendingRevoke = null;
	}}
/>
