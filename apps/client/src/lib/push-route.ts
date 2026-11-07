/** Maps a server notification path such as "sessions/abc" to an app route. */
export function routeFor(path: string | undefined): string {
	const match = /^sessions\/([^/?#]+)$/.exec(path ?? '');
	return match ? `/s/${match[1]}` : '/sessions';
}

export type OpenWindow = { url: string; focused: boolean; visibilityState?: DocumentVisibilityState };

/**
 * Says whether a push needs a system notification. A focused, visible app window
 * already gets the same notification from the live stream, which stays silent on
 * the conversation it leads to and shows a toast on any other route, so a system
 * notification there would arrive twice.
 */
export function needsSystemNotification(windows: OpenWindow[], origin: string): boolean {
	return !windows.some(
		(w) => w.focused && w.visibilityState !== 'hidden' && new URL(w.url, origin).origin === origin
	);
}
