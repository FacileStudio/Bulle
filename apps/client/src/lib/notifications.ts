import { goto } from '$app/navigation';
import { PushNotifications } from '@capacitor/push-notifications';
import { backend, type PushConfig } from '$lib/backend';
import { isNative } from '$lib/platform';
import { routeFor } from '$lib/push-route';

export type Permission = 'granted' | 'denied' | 'prompt' | 'unsupported';

let listening = false;

/** Reads whether this device lets Bulle show notifications. */
export async function permission(): Promise<Permission> {
	if (isNative()) {
		const { receive } = await PushNotifications.checkPermissions();
		return receive === 'granted' ? 'granted' : receive === 'denied' ? 'denied' : 'prompt';
	}
	if (!('Notification' in window) || !('serviceWorker' in navigator) || !('PushManager' in window)) {
		return 'unsupported';
	}
	const state = Notification.permission;
	return state === 'default' ? 'prompt' : state;
}

/** Says whether the server can deliver to this kind of device. */
export function supported(config: PushConfig | null): boolean {
	if (!config) return false;
	return isNative() ? config.native : !!config.web_push_key;
}

function listenNative() {
	if (listening) return;
	listening = true;
	void PushNotifications.addListener('registration', ({ value }) => {
		void backend.registerDevice({ kind: 'fcm', token: value });
	});
	void PushNotifications.addListener('pushNotificationActionPerformed', ({ notification }) => {
		void goto(routeFor(notification.data?.path));
	});
}

function base64ToBytes(base64: string): Uint8Array<ArrayBuffer> {
	const padded = (base64 + '='.repeat((4 - (base64.length % 4)) % 4)).replace(/-/g, '+').replace(/_/g, '/');
	const raw = atob(padded);
	const bytes = new Uint8Array(new ArrayBuffer(raw.length));
	for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
	return bytes;
}

async function subscribeWeb(publicKey: string) {
	const registration = await navigator.serviceWorker.ready;
	const subscription =
		(await registration.pushManager.getSubscription()) ??
		(await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: base64ToBytes(publicKey)
		}));
	const keys = subscription.toJSON().keys ?? {};
	await backend.registerDevice({
		kind: 'web',
		token: subscription.endpoint,
		p256dh: keys.p256dh,
		auth: keys.auth
	});
}

/**
 * Registers this device with the server when permission is already granted, so a
 * rotated FCM token or a fresh browser subscription keeps working. Never prompts.
 */
export async function syncDevice(config: PushConfig | null): Promise<void> {
	if (!supported(config) || (await permission()) !== 'granted') return;
	if (isNative()) {
		listenNative();
		await PushNotifications.register();
		return;
	}
	await subscribeWeb(config!.web_push_key!);
}

/** Asks the system for permission, then registers this device. Call it from a click. */
export async function enable(config: PushConfig | null): Promise<Permission> {
	if (!supported(config)) return 'unsupported';
	if (isNative()) {
		listenNative();
		const { receive } = await PushNotifications.requestPermissions();
		if (receive !== 'granted') return 'denied';
		await PushNotifications.register();
		return 'granted';
	}
	const result = await Notification.requestPermission();
	if (result !== 'granted') return result === 'denied' ? 'denied' : 'prompt';
	await subscribeWeb(config!.web_push_key!);
	return 'granted';
}
