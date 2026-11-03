export const mobileSidebar = $state({ open: false });

const COLLAPSED_KEY = 'bulle.sidebar.collapsed';
const WIDE = '(min-width: 768px)';

function storedCollapsed(): boolean {
	try {
		return localStorage.getItem(COLLAPSED_KEY) === '1';
	} catch {
		return false;
	}
}

export const desktopSidebar = $state({ collapsed: typeof window !== 'undefined' && storedCollapsed() });

/** Collapses or expands the docked sidebar shown on wide screens, and remembers the choice. */
export function setCollapsed(collapsed: boolean) {
	desktopSidebar.collapsed = collapsed;
	try {
		localStorage.setItem(COLLAPSED_KEY, collapsed ? '1' : '0');
	} catch {
		return;
	}
}

/** Shows the sidebar: expands the docked one on wide screens, opens the drawer on narrow ones. */
export function openSidebar() {
	if (window.matchMedia(WIDE).matches) setCollapsed(false);
	else mobileSidebar.open = true;
}

/** Reports whether a key press is the Ctrl+B or Cmd+B sidebar toggle. */
export function isSidebarShortcut(event: Pick<KeyboardEvent, 'key' | 'ctrlKey' | 'metaKey' | 'altKey' | 'shiftKey'>): boolean {
	return (event.ctrlKey || event.metaKey) && !event.altKey && !event.shiftKey && event.key.toLowerCase() === 'b';
}

const DECIDE = 12;
const COMMIT = 60;
const NO_SWIPE = 'pre, code, textarea, input, select, [data-no-swipe], .overflow-x-auto';

/**
 * Reads a finished horizontal touch and says whether it should open or close the sidebar.
 * Returns null when the gesture was vertical, too short, or pointed the wrong way.
 */
export function swipeIntent(dx: number, dy: number, open: boolean): 'open' | 'close' | null {
	if (Math.abs(dx) < COMMIT || Math.abs(dx) < Math.abs(dy) * 1.5) return null;
	if (dx > 0 && !open) return 'open';
	if (dx < 0 && open) return 'close';
	return null;
}

/**
 * Listens for horizontal swipes on the window: right opens the sidebar, left closes it.
 * Touches that start inside code, form fields or horizontal scrollers are ignored.
 */
export function listenForSwipes(): () => void {
	let startX = 0;
	let startY = 0;
	let tracking = false;

	function onstart(event: TouchEvent) {
		const touch = event.touches[0];
		const target = event.target instanceof Element ? event.target : null;
		tracking =
			event.touches.length === 1 &&
			!window.matchMedia('(min-width: 768px)').matches &&
			!target?.closest(NO_SWIPE);
		startX = touch.clientX;
		startY = touch.clientY;
	}

	function onend(event: TouchEvent) {
		if (!tracking) return;
		tracking = false;
		const touch = event.changedTouches[0];
		const dx = touch.clientX - startX;
		const dy = touch.clientY - startY;
		if (Math.abs(dx) < DECIDE) return;
		const intent = swipeIntent(dx, dy, mobileSidebar.open);
		if (intent) mobileSidebar.open = intent === 'open';
	}

	window.addEventListener('touchstart', onstart, { passive: true });
	window.addEventListener('touchend', onend, { passive: true });
	return () => {
		window.removeEventListener('touchstart', onstart);
		window.removeEventListener('touchend', onend);
	};
}
