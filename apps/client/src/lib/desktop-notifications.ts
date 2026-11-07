import {
	isPermissionGranted,
	onAction,
	requestPermission,
	sendNotification
} from '@tauri-apps/plugin-notification';
import { getCurrentWindow } from '@tauri-apps/api/window';
import { goto } from '$app/navigation';
import { showNotifications } from '$lib/live/sync';

const routes = new Map<number, string>();
let nextId = 1;
const MAX_ROUTES = 50;
let clicks: Promise<unknown> | null = null;

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

/** Opens the route of a clicked desktop notification and brings the window forward. */
function listenClicks(): void {
	clicks ??= onAction(({ notification }) => {
		if (!notification) return;
		const extra = notification.extra?.route;
		const route = typeof extra === 'string' ? extra : routes.get(notification.id);
		routes.delete(notification.id);
		if (!route) return;
		void getCurrentWindow().setFocus().catch(() => undefined);
		void goto(route);
	}).catch(() => null);
}

/** Shows the live stream's notifications as desktop OS notifications that open their route on click. */
export function listenDesktop(): void {
	listenClicks();
	showNotifications((note, route) => {
		const id = nextId++;
		routes.set(id, route);
		if (routes.size > MAX_ROUTES) routes.delete(routes.keys().next().value!);
		sendNotification({ id, title: note.title || 'Bulle', body: note.body, extra: { route } });
	});
}
