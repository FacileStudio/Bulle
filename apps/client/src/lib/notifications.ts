import { goto } from '$app/navigation';
import { PushNotifications } from '@capacitor/push-notifications';
import { backend, type PushConfig } from '$lib/backend';
import { desktopPermission, enableDesktop, listenDesktop } from '$lib/desktop-notifications';
import { webDelivery } from '$lib/live/delivery.svelte';
import type { LiveNotification } from '$lib/live/events';
import { showNotifications } from '$lib/live/sync';
import { isDesktop, isNative } from '$lib/platform';
import { routeFor } from '$lib/push-route';

export type Permission = 'granted' | 'denied' | 'prompt' | 'unsupported';

const CHANNEL_ID = 'bulle';

let listening = false;
let nativeToken: string | null = null;

/** Reads whether this device lets Bulle show notifications. */
export async function permission(): Promise<Permission> {
	if (isDesktop()) return desktopPermission();
	if (isNative()) {
		const { receive } = await PushNotifications.checkPermissions();
		return receive === 'granted' ? 'granted' : receive === 'denied' ? 'denied' : 'prompt';
	}
	if (!('Notification' in window)) return 'unsupported';
	const state = Notification.permission;
	return state === 'default' ? 'prompt' : state;
}

/** Says whether the server can deliver to this kind of device. */
export function supported(config: PushConfig | null): boolean {
	if (!config) return false;
	if (isDesktop()) return true;
	return isNative() ? config.native : !!config.web_push_key;
}

function listenNative() {
	if (listening) return;
	listening = true;
	void PushNotifications.addListener('registration', ({ value }) => {
		nativeToken = value;
		void backend.registerDevice({ kind: 'fcm', token: value });
	});
	void PushNotifications.addListener('registrationError', ({ error }) => {
		console.warn('push registration failed', error);
	});
	void PushNotifications.addListener('pushNotificationActionPerformed', ({ notification }) => {
		void goto(routeFor(notification.data?.path));
	});
}

/**
 * Registers with FCM on a high-importance channel. The server sends to the same channel id
 * (push.ChannelID); without it Android files Bulle under a silent "Miscellaneous" channel.
 */
async function registerNative(): Promise<void> {
	listenNative();
	await PushNotifications.createChannel({
		id: CHANNEL_ID,
		name: 'Bulle',
		description: 'Approvals and finished jobs',
		importance: 4,
		visibility: 0
	});
	await PushNotifications.register();
}

function base64ToBytes(base64: string): Uint8Array<ArrayBuffer> {
	const padded = (base64 + '='.repeat((4 - (base64.length % 4)) % 4)).replace(/-/g, '+').replace(/_/g, '/');
	const raw = atob(padded);
	const bytes = new Uint8Array(new ArrayBuffer(raw.length));
	for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
	return bytes;
}

export const PUSH_SERVICE_OFF =
	'This browser has no push service turned on. In Brave, enable "Use Google services for push messaging" in brave://settings/privacy and restart it. Helium and other Chromium builds without Google services cannot receive web push: use Chrome, Firefox or the Bulle app instead.';

export const TAB_ONLY =
	'This browser has no push service, so Bulle notifies you only while a tab is open. Install the desktop app for notifications when it is closed.';

function browserNotify(note: LiveNotification, route: string): void {
	const shown = new Notification(note.title || 'Bulle', { body: note.body, tag: note.path || undefined });
	shown.onclick = () => {
		window.focus();
		shown.close();
		void goto(route);
	};
}

function webPushReady(config: PushConfig | null): boolean {
	return 'serviceWorker' in navigator && 'PushManager' in window && !!config?.web_push_key;
}

async function deliverWeb(config: PushConfig | null): Promise<void> {
	if (webPushReady(config)) {
		try {
			await subscribeWeb(config!.web_push_key!);
			webDelivery.tabOnly = false;
			return;
		} catch (err) {
			if (!(err instanceof Error && err.message === PUSH_SERVICE_OFF)) throw err;
		}
	}
	webDelivery.tabOnly = true;
	showNotifications(browserNotify);
}

async function pushSubscription(registration: ServiceWorkerRegistration, publicKey: string) {
	try {
		return await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: base64ToBytes(publicKey)
		});
	} catch (err) {
		if (err instanceof DOMException && err.name === 'AbortError') throw new Error(PUSH_SERVICE_OFF);
		throw err;
	}
}

async function subscribeWeb(publicKey: string) {
	const registration = await navigator.serviceWorker.ready;
	const subscription =
		(await registration.pushManager.getSubscription()) ?? (await pushSubscription(registration, publicKey));
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
	if (isDesktop()) {
		listenDesktop();
		return;
	}
	if (isNative()) {
		await registerNative();
		return;
	}
	await deliverWeb(config);
}

/** Asks the system for permission, then registers this device. Call it from a click. */
export async function enable(config: PushConfig | null): Promise<Permission> {
	if (!supported(config)) return 'unsupported';
	if (isDesktop()) return enableDesktop();
	if (isNative()) {
		const { receive } = await PushNotifications.requestPermissions();
		if (receive !== 'granted') return 'denied';
		await registerNative();
		return 'granted';
	}
	const result = await Notification.requestPermission();
	if (result !== 'granted') return result === 'denied' ? 'denied' : 'prompt';
	await deliverWeb(config);
	return 'granted';
}

/**
 * Stops this device from receiving the signed-out user's notifications: the server forgets
 * its token, and the phone or browser drops its push registration. Call it while the session
 * still works. Best effort: a failure never blocks signing out.
 */
export async function forgetDevice(): Promise<void> {
	try {
		if (isDesktop()) return;
		if (isNative()) {
			if (nativeToken) await backend.forgetDevice(nativeToken);
			nativeToken = null;
			await PushNotifications.unregister();
			return;
		}
		if (!('serviceWorker' in navigator)) return;
		const registration = await navigator.serviceWorker.getRegistration();
		const subscription = await registration?.pushManager.getSubscription();
		if (!subscription) return;
		await backend.forgetDevice(subscription.endpoint);
		await subscription.unsubscribe();
	} catch (err) {
		console.warn('could not forget this push device', err);
	}
}
