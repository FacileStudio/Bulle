import { t } from '$lib/i18n/index.svelte';

export const SUBSCRIBE_URL = 'https://sso.facile.studio/spaces';

export function planLabel(plan: 'free' | 'paid' | undefined): string | null {
	if (!plan) return null;
	return t(plan === 'paid' ? 'account.plan.paid' : 'account.plan.free');
}
