import { goto } from '$app/navigation';
import { toast } from '@facile/muse';
import { chat } from '$lib/chat/chat.svelte';
import { sessions } from '$lib/chat/sessions.svelte';
import { refreshJobs } from '$lib/jobs-activity.svelte';
import { routeFor } from '$lib/push-route';
import type { LiveEvent, LiveNotification } from './events';
import { live } from './live';
import { noteAction } from './notify';

export type SystemNotifier = (note: LiveNotification, route: string) => void;

let notifier: SystemNotifier | null = null;

function isCronRun(event: LiveEvent): boolean {
	return (event.type === 'run.started' || event.type === 'run.finished') && event.data.channel === 'cron';
}

function notify(note: LiveNotification): void {
	if (!notifier) return;
	const route = routeFor(note.path);
	const action = noteAction(route, {
		openRoute: chat.current ? `/s/${chat.current}` : null,
		visible: document.visibilityState === 'visible',
		focused: document.hasFocus()
	});
	if (action === 'toast') {
		toast.info(note.body, { action: { label: 'Open', onClick: () => void goto(route) } });
	} else if (action === 'system') {
		notifier(note, route);
	}
}

/** Connects this app to the live stream and keeps sessions, the open chat and jobs in sync. */
export function startLive(): () => void {
	const off = live.listen({
		onOpen: () => {
			void sessions.resync();
			chat.resync();
		},
		onEvent: (event) => {
			if (event.type === 'notification') {
				notify(event.data);
				return;
			}
			sessions.apply(event);
			chat.applyLive(event);
			if (isCronRun(event)) void refreshJobs();
		}
	});
	live.start();
	return () => {
		off();
		live.stop();
	};
}

/** Shows the user's notifications from the live stream with this system notifier. */
export function showNotifications(system: SystemNotifier): void {
	notifier = system;
}
