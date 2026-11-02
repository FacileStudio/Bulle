/// <reference lib="webworker" />

import { routeFor } from '$lib/push-route';

const worker = self as unknown as ServiceWorkerGlobalScope;

type Payload = { title?: string; body?: string; path?: string };

worker.addEventListener('install', () => void worker.skipWaiting());
worker.addEventListener('activate', (event) => event.waitUntil(worker.clients.claim()));

worker.addEventListener('push', (event) => {
	let data: Payload = {};
	try {
		data = event.data?.json() ?? {};
	} catch {
		data = { body: event.data?.text() };
	}
	event.waitUntil(
		worker.registration.showNotification(data.title || 'Bulle', {
			body: data.body,
			data: { url: routeFor(data.path) }
		})
	);
});

worker.addEventListener('notificationclick', (event) => {
	event.notification.close();
	const url = new URL(event.notification.data?.url ?? '/sessions', worker.location.origin).href;
	event.waitUntil(
		worker.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windows) => {
			const open = windows.find((w) => new URL(w.url).origin === worker.location.origin);
			if (open) return open.focus().then(() => open.navigate(url));
			return worker.clients.openWindow(url);
		})
	);
});
