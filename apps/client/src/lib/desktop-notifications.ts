import {
	isPermissionGranted,
	requestPermission,
	sendNotification
} from '@tauri-apps/plugin-notification';
import { showNotifications } from '$lib/live/sync';

/** Reads whether the desktop OS lets Bulle show notifications. */
export async function desktopPermission(): Promise<'granted' | 'prompt'> {
	return (await isPermissionGranted()) ? 'granted' : 'prompt';
}

/** Asks the desktop OS for permission, then starts listening. */
export async function enableDesktop(): Promise<'granted' | 'denied'> {
	if (!(await isPermissionGranted()) && (await requestPermission()) !== 'granted') return 'denied';
	listenDesktop();
	return 'granted';
}

/** Shows the live stream's notifications as desktop OS notifications. */
export function listenDesktop(): void {
	showNotifications((note) => sendNotification({ title: note.title || 'Bulle', body: note.body }));
}
