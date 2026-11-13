import { t } from '$lib/i18n/index.svelte';
import { App } from '@capacitor/app';
import { Browser } from '@capacitor/browser';
import { SecureStoragePlugin } from 'capacitor-secure-storage-plugin';
import { getCurrentWindow } from '@tauri-apps/api/window';
import { getCurrent, onOpenUrl } from '@tauri-apps/plugin-deep-link';
import { openUrl } from '@tauri-apps/plugin-opener';
import { goto } from '$app/navigation';
import {
	apiBase,
	isDesktop,
	isNative,
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

/**
 * Keeps the pending login where it survives the trip through the browser: the secure store
 * on Android, where the OS may kill the app meanwhile, and sessionStorage elsewhere.
 */
const store = {
	async get(): Promise<string | null> {
		if (isNative()) return (await SecureStoragePlugin.get({ key: PENDING_KEY })).value;
		return sessionStorage.getItem(PENDING_KEY);
	},
	async set(value: string): Promise<void> {
		if (isNative()) await SecureStoragePlugin.set({ key: PENDING_KEY, value });
		else sessionStorage.setItem(PENDING_KEY, value);
	},
	async remove(): Promise<void> {
		if (isNative()) await SecureStoragePlugin.remove({ key: PENDING_KEY });
		else sessionStorage.removeItem(PENDING_KEY);
	}
};

async function remember(value: Pending) {
	pending = value;
	await store.set(JSON.stringify(value)).catch(() => undefined);
}

async function recall(): Promise<Pending | null> {
	if (pending) return pending;
	try {
		const raw = await store.get();
		return raw ? (JSON.parse(raw) as Pending) : null;
	} catch {
		return null;
	}
}

async function forget() {
	pending = null;
	await store.remove().catch(() => undefined);
}

export async function startNativeLogin(): Promise<void> {
	const verifier = randomString(64);
	const state = randomState();
	await remember({ verifier, state });
	const challenge = await pkceChallenge(verifier);
	const query = new URLSearchParams({ flow: 'mobile', challenge, state });
	const url = `${apiBase()}/auth/oidc?${query}`;
	if (isDesktop()) await openUrl(url);
	else await Browser.open({ url });
}

async function exchange(code: string, state: string): Promise<void> {
	const saved = await recall();
	await forget();
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

const handled = new Set<string>();

function handleDeepLink(url: string) {
	if (handled.has(url)) return;
	handled.add(url);
	const link = parseDeepLink(url);
	if (link?.kind === 'auth') void finishLogin(link.code, link.state, link.error);
	else if (link?.kind === 'session') void goto(`/s/${encodeURIComponent(link.id)}`);
}

/** Reads the deep link that launched the app, when the platform kept one. */
async function launchUrls(): Promise<string[]> {
	try {
		if (isDesktop()) return (await getCurrent()) ?? [];
		const launch = await App.getLaunchUrl();
		return launch?.url ? [launch.url] : [];
	} catch {
		return [];
	}
}

export function registerDeepLinks(): void {
	if (registered) return;
	registered = true;
	if (isDesktop()) void onOpenUrl((urls) => urls.forEach(handleDeepLink));
	else void App.addListener('appUrlOpen', ({ url }) => handleDeepLink(url));
	void launchUrls().then(async (urls) => {
		for (const url of urls) {
			if (parseDeepLink(url)?.kind === 'auth' && !(await recall())) continue;
			handleDeepLink(url);
		}
	});
}
