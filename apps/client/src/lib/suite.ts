import { ApiError, apiFetch } from '$lib/backend';

export type SuiteState = 'connected' | 'no_account' | 'reconnect' | 'unavailable';

export type SuiteApp = {
	app: string;
	name: string;
	url: string;
	state: SuiteState;
};

export type SuiteBadge = { tone: 'success' | 'warning' | 'danger' | 'neutral'; label: string };

const BADGES: Record<SuiteState, SuiteBadge> = {
	connected: { tone: 'success', label: 'Connected' },
	no_account: { tone: 'warning', label: 'No account yet' },
	reconnect: { tone: 'danger', label: 'Reconnect' },
	unavailable: { tone: 'neutral', label: 'Unavailable' }
};

export function badgeFor(state: SuiteState): SuiteBadge {
	return BADGES[state] ?? BADGES.unavailable;
}

export async function getSuiteStatus(): Promise<SuiteApp[]> {
	try {
		const body = await apiFetch<{ apps: SuiteApp[] }>('/suite/status');
		return body.apps ?? [];
	} catch (err) {
		if (err instanceof ApiError && err.status === 404) return [];
		throw err;
	}
}
