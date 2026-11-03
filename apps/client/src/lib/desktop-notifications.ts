import { goto } from '$app/navigation';
import { toast } from '@facile/muse';
import {
	isPermissionGranted,
	requestPermission,
	sendNotification
} from '@tauri-apps/plugin-notification';
import { authedInit, backend } from '$lib/backend';
import { routeFor } from '$lib/push-route';
import { createSseParser } from '$lib/sse';

type StreamNotification = { title: string; body: string; path: string };

const MAX_BACKOFF_MS = 30000;

let listening = false;

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

function show(note: StreamNotification) {
	const route = routeFor(note.path);
	if (document.hasFocus()) {
		toast.info(note.body, { action: { label: 'Open', onClick: () => void goto(route) } });
		return;
	}
	sendNotification({ title: note.title || 'Bulle', body: note.body });
}

async function readStream(): Promise<boolean> {
	const init = await authedInit({ headers: { Accept: 'text/event-stream' } });
	const response = await fetch(`${backend.baseUrl}/push/stream`, init);
	if (!response.ok || !response.body) return false;
	const parse = createSseParser((frame) => {
		if (frame.event !== 'notification') return;
		try {
			show(JSON.parse(frame.data) as StreamNotification);
		} catch {
			return;
		}
	});
	const decoder = new TextDecoder();
	const reader = response.body.getReader();
	for (;;) {
		const { done, value } = await reader.read();
		if (done) return true;
		parse(decoder.decode(value, { stream: true }));
	}
}

/** Keeps one notification stream open for this window, reconnecting with backoff. */
export function listenDesktop(): void {
	if (listening) return;
	listening = true;
	let backoff = 1000;
	const loop = async () => {
		for (;;) {
			const opened = await readStream().catch(() => false);
			backoff = opened ? 1000 : Math.min(backoff * 2, MAX_BACKOFF_MS);
			await new Promise((resolve) => setTimeout(resolve, backoff));
		}
	};
	void loop();
}
