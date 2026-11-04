<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Alert,
		Badge,
		Button,
		Card,
		ConfirmModal,
		Input,
		Section,
		Spinner,
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
	import { isApp } from '$lib/platform';

	let devices = $state<HardwareDevice[]>([]);
	let loadError = $state('');
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
			loadError = err instanceof Error ? err.message : 'Could not load your devices';
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
			toast.success(`${paired.device.name} paired`);
			await load();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not pair the device');
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
			toast.danger(err instanceof Error ? err.message : 'Could not connect the device');
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
			toast.success(`${device.name} revoked`);
			await load();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not revoke the device');
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
	title="Devices"
	description="Pastille and other BOHS-01 devices paired to your account. Bulle relays a paired device over Bluetooth while this page is open; with Wi-Fi set, it reaches Bulle on its own."
>
	{#snippet actions()}
		<Button icon={icons.plus} disabled={!supported || pairing} onclick={pair}>
			{#if pairing}<Spinner size="sm" />{/if}
			Pair device
		</Button>
	{/snippet}
	<div class="flex flex-col gap-4">
		{#if !supported}
			<Alert tone="warning">
				{isApp()
					? 'Pair and relay devices from Bulle in Chrome or Edge on the web.'
					: 'This browser has no Web Bluetooth. Open Bulle in Chrome or Edge to pair a device.'}
			</Alert>
		{/if}
		{#if loadError}
			<Alert tone="danger">{loadError}</Alert>
		{/if}
		{#if progress}
			<Alert tone="info">{progress}</Alert>
		{/if}
		<div class="flex flex-col gap-2 sm:flex-row">
			<Input
				bind:value={wifiSsid}
				autocomplete="off"
				aria-label="Wi-Fi network (optional)"
				placeholder="Wi-Fi network (optional)"
			/>
			<Input
				type="password"
				bind:value={wifiPsk}
				autocomplete="off"
				aria-label="Wi-Fi password"
				placeholder="Wi-Fi password"
			/>
		</div>
		<div class="flex flex-col gap-2">
			{#each devices as device (device.id)}
				<Card class="flex items-center gap-3">
					<div class="flex min-w-0 flex-1 flex-col">
						<span class="truncate font-medium text-fc-fg">{device.name}</span>
						<span class="text-fc-xs text-fc-fg-muted">
							{device.last_seen_at ? `Seen ${relativeTime(device.last_seen_at)}` : 'Never connected'}
						</span>
					</div>
					{#if device.last_transport}
						<Badge>{device.last_transport === 'ble' ? 'Bluetooth' : 'Wi-Fi'}</Badge>
					{/if}
					<Button
						variant="ghost"
						icon={icons.revoke}
						onclick={() => {
							pendingRevoke = device;
							confirmOpen = true;
						}}
					>
						Revoke
					</Button>
				</Card>
			{:else}
				<span class="text-fc-sm text-fc-fg-muted">No device paired yet.</span>
			{/each}
		</div>
	</div>
</Section>

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title={`Revoke ${pendingRevoke?.name ?? 'this device'}?`}
	description="Its token stops working at once. Pair it again to use it."
	confirmLabel="Revoke device"
	onConfirm={confirmRevoke}
	onCancel={() => {
		confirmOpen = false;
		pendingRevoke = null;
	}}
/>
