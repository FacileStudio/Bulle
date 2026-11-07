import { t } from '$lib/i18n/index.svelte';
import { App } from '@capacitor/app';
import { Browser } from '@capacitor/browser';
import { getCurrentWindow } from '@tauri-apps/api/window';
import { onOpenUrl } from '@tauri-apps/plugin-deep-link';
import { openUrl } from '@tauri-apps/plugin-opener';
import { goto } from '$app/navigation';
import {
	apiBase,
	isDesktop,
	parseDeepLink,
	pkceChallenge,
	randomState,
	randomString,
	setToken
} from '$lib/platform';
import { HOME } from '$lib/nav';
import { loadSession } from '$lib/session.svelte';

const PENDING_KEY = 'bulle.pending-login';

type Pending = { verifier: string; state: string };

let pending: Pending | null = null;
let registered = false;

function remember(value: Pending) {
	pending = value;
	try {
		sessionStorage.setItem(PENDING_KEY, JSON.stringify(value));
	} catch {
		return;
	}
}

function recall(): Pending | null {
	if (pending) return pending;
	try {
		const raw = sessionStorage.getItem(PENDING_KEY);
		return raw ? (JSON.parse(raw) as Pending) : null;
	} catch {
		return null;
	}
}

function forget() {
	pending = null;
	try {
		sessionStorage.removeItem(PENDING_KEY);
	} catch {
		return;
	}
}

export async function startNativeLogin(): Promise<void> {
	const verifier = randomString(64);
	const state = randomState();
	remember({ verifier, state });
	const challenge = await pkceChallenge(verifier);
	const query = new URLSearchParams({ flow: 'mobile', challenge, state });
	const url = `${apiBase()}/auth/oidc?${query}`;
	if (isDesktop()) await openUrl(url);
	else await Browser.open({ url });
}

async function exchange(code: string, state: string): Promise<void> {
	const saved = recall();
	forget();
	if (!saved || saved.state !== state) throw new Error(t('native.error.state'));
	const response = await fetch(`${apiBase()}/auth/oidc/exchange`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ code, verifier: saved.verifier })
	});
	if (!response.ok) throw new Error(t('native.error.incomplete'));
	const body = (await response.json()) as { token?: string };
	if (!body.token) throw new Error(t('native.error.incomplete'));
	await setToken(body.token);
}

async function finishLogin(code: string, state: string, error: string) {
	let failure = error;
	if (!failure) {
		try {
			await exchange(code, state);
		} catch (err) {
			failure = err instanceof Error ? err.message : t('native.error.failed');
		}
	}
	if (isDesktop()) await getCurrentWindow().setFocus().catch(() => undefined);
	else await Browser.close().catch(() => undefined);
	if (failure) {
		await goto(`/login?error=${encodeURIComponent(failure)}`, { replaceState: true });
		return;
	}
	await loadSession();
	await goto(HOME, { replaceState: true });
}

function handleDeepLink(url: string) {
	const link = parseDeepLink(url);
	if (link?.kind === 'auth') void finishLogin(link.code, link.state, link.error);
	else if (link?.kind === 'session') void goto(`/s/${encodeURIComponent(link.id)}`);
}

export function registerDeepLinks(): void {
	if (registered) return;
	registered = true;
	if (isDesktop()) void onOpenUrl((urls) => urls.forEach(handleDeepLink));
	else void App.addListener('appUrlOpen', ({ url }) => handleDeepLink(url));
}
