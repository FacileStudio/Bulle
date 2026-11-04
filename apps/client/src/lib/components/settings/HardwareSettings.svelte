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
	import { bluetoothAvailable, scanAndPairBohsDevice, type PairedDevice } from '$lib/hardware/ble';

	let devices = $state<HardwareDevice[]>([]);
	let loadError = $state('');
	let pairing = $state(false);
	let progress = $state('');
	let wifiSsid = $state('');
	let wifiPsk = $state('');
	let relays = $state.raw<PairedDevice[]>([]);
	let pendingRevoke = $state<HardwareDevice | null>(null);
	let confirmOpen = $state(false);
	const supported = bluetoothAvailable();

	/** Reloads the user's paired devices. */
	async function load() {
		try {
			devices = await listDevices();
			loadError = '';
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Could not load your devices';
		}
	}

	/** Pairs a Pastille over Bluetooth and keeps relaying it while this page is open. */
	async function pair() {
		pairing = true;
		try {
			const paired = await scanAndPairBohsDevice({
				wifiSsid: wifiSsid.trim() || undefined,
				wifiPsk: wifiPsk || undefined,
				onStatus: (message) => (progress = message)
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
			<Alert tone="warning">This browser has no Web Bluetooth. Open Bulle in Chrome or Edge to pair a device.</Alert>
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
