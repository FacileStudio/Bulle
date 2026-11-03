export type NoteView = { openRoute: string | null; visible: boolean; focused: boolean };

export type NoteAction = 'skip' | 'toast' | 'system';

/**
 * Decides how to show a notification that leads to route: not at all when the user
 * is already looking at that conversation, as a toast when the app has focus, and
 * as a system notification otherwise.
 */
export function noteAction(route: string, view: NoteView): NoteAction {
	if (view.visible && view.openRoute === route) return 'skip';
	if (view.visible && view.focused) return 'toast';
	return 'system';
}
