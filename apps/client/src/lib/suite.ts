import { ApiError, apiFetch } from '$lib/backend';
import { t } from '$lib/i18n/index.svelte';

export type SuiteState = 'connected' | 'no_account' | 'reconnect' | 'unavailable';

export type SuiteApp = {
	app: string;
	name: string;
	url: string;
	state: SuiteState;
};

export type SuiteBadge = { tone: 'success' | 'warning' | 'danger' | 'neutral'; label: string };

const TONES: Record<SuiteState, SuiteBadge['tone']> = {
	connected: 'success',
	no_account: 'warning',
	reconnect: 'danger',
	unavailable: 'neutral'
};

export function badgeFor(state: SuiteState): SuiteBadge {
	const known: SuiteState = state in TONES ? state : 'unavailable';
	return { tone: TONES[known], label: t(`suite.state.${known}`) };
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
