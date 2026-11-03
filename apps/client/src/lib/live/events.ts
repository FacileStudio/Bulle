import type { SessionStatus, SessionSummary } from '$lib/api';
import type { SseFrame } from '$lib/sse';

export type LiveSession = {
	session_id: string;
	space_id: string;
	title: string;
	channel?: string;
	updated_at: string;
	status: SessionStatus;
};

export type LiveRun = {
	session_id: string;
	run_id: string;
	channel?: string;
	status?: string;
	session_status: SessionStatus;
};

export type LiveApproval = {
	session_id: string;
	run_id: string;
	call_id: string;
	tool?: string;
	approved?: boolean;
	session_status: SessionStatus;
};

export type LiveNotification = { title: string; body: string; path: string };

export type LiveEvent =
	| { type: 'session.created' | 'session.updated'; data: LiveSession }
	| { type: 'session.deleted'; data: { session_id: string } }
	| { type: 'run.started' | 'run.finished'; data: LiveRun }
	| { type: 'approval.pending' | 'approval.decided'; data: LiveApproval }
	| { type: 'notification'; data: LiveNotification };

const TYPES = new Set<string>([
	'session.created',
	'session.updated',
	'session.deleted',
	'run.started',
	'run.finished',
	'approval.pending',
	'approval.decided',
	'notification'
]);

/** Reads one server-sent frame of the live stream, or null for a frame this client does not know. */
export function parseLiveFrame(frame: SseFrame): LiveEvent | null {
	if (!TYPES.has(frame.event)) return null;
	try {
		return { type: frame.event, data: JSON.parse(frame.data) } as LiveEvent;
	} catch {
		return null;
	}
}

function byRecent(items: SessionSummary[]): SessionSummary[] {
	return [...items].sort((a, b) => Date.parse(b.updated_at) - Date.parse(a.updated_at));
}

function upsert(items: SessionSummary[], data: LiveSession, spaceId: string | null): SessionSummary[] {
	const rest = items.filter((s) => s.id !== data.session_id);
	if (spaceId && data.space_id !== spaceId) return rest;
	const previous = items.find((s) => s.id === data.session_id);
	const next: SessionSummary = {
		...previous,
		id: data.session_id,
		space_id: data.space_id,
		title: data.title,
		channel: data.channel,
		updated_at: data.updated_at,
		status: data.status
	};
	return byRecent([...rest, next]);
}

function withStatus(items: SessionSummary[], id: string, status: SessionStatus): SessionSummary[] {
	if (!items.some((s) => s.id === id && s.status !== status)) return items;
	return items.map((s) => (s.id === id ? { ...s, status } : s));
}

/**
 * Applies one live event to the session list of the space being shown (null shows
 * every space). Returns the same array when the event changes nothing.
 */
export function applyLiveEvent(
	items: SessionSummary[],
	event: LiveEvent,
	spaceId: string | null
): SessionSummary[] {
	switch (event.type) {
		case 'session.created':
		case 'session.updated':
			return upsert(items, event.data, spaceId);
		case 'session.deleted':
			return items.some((s) => s.id === event.data.session_id)
				? items.filter((s) => s.id !== event.data.session_id)
				: items;
		case 'run.started':
		case 'run.finished':
		case 'approval.pending':
		case 'approval.decided':
			return withStatus(items, event.data.session_id, event.data.session_status);
		default:
			return items;
	}
}
