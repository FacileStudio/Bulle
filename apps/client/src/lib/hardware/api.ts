import { apiFetch, backend } from '$lib/backend';

export type HardwareDevice = {
	id: string;
	name: string;
	device_class: 'charm' | 'field' | 'totem';
	hardware_id: string;
	capabilities: unknown[];
	last_seen_at?: string;
	last_transport?: 'ble' | 'wifi' | 'usb';
	created_at: string;
};

export type ClaimedDevice = {
	device: HardwareDevice;
	device_token: string;
	gateway_ws_url: string;
};

/** Lists the signed-in user's paired BOHS-01 devices. */
export function listDevices(): Promise<HardwareDevice[]> {
	return apiFetch<HardwareDevice[]>('/hardware/devices');
}

/** Pairs a device to the signed-in user. The answer holds the only copy of its token. */
export function claimDevice(hardwareId: string, name?: string): Promise<ClaimedDevice> {
	return apiFetch<ClaimedDevice>('/hardware/devices/claim', {
		method: 'POST',
		body: JSON.stringify({ hardware_id: hardwareId, ...(name ? { name } : {}) })
	});
}

/** Revokes one of the user's devices; the gateway closes its live sockets. */
export function revokeDevice(id: string): Promise<void> {
	return apiFetch<void>(`/hardware/devices/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/** Answers the relay socket URL for one of the user's devices. */
export function relayUrl(deviceId: string): string {
	const url = new URL(`${backend.baseUrl}/hardware/ws`, window.location.href);
	url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
	url.searchParams.set('device_id', deviceId);
	return url.toString();
}
