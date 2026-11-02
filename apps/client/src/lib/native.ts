import { App } from '@capacitor/app';
import { Browser } from '@capacitor/browser';
import { goto } from '$app/navigation';
import { apiBase, parseDeepLink, pkceChallenge, randomString, setToken } from '$lib/platform';
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
	const state = randomString(32);
	remember({ verifier, state });
	const challenge = await pkceChallenge(verifier);
	const query = new URLSearchParams({ flow: 'mobile', challenge, state });
	await Browser.open({ url: `${apiBase()}/auth/oidc?${query}` });
}

async function exchange(code: string, state: string): Promise<void> {
	const saved = recall();
	forget();
	if (!saved || saved.state !== state) throw new Error('Sign-in state did not match. Try again.');
	const response = await fetch(`${apiBase()}/auth/oidc/exchange`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ code, verifier: saved.verifier })
	});
	if (!response.ok) throw new Error('Sign-in could not be completed. Try again.');
	const body = (await response.json()) as { token?: string };
	if (!body.token) throw new Error('Sign-in could not be completed. Try again.');
	await setToken(body.token);
}

async function finishLogin(code: string, state: string, error: string) {
	let failure = error;
	if (!failure) {
		try {
			await exchange(code, state);
		} catch (err) {
			failure = err instanceof Error ? err.message : 'Sign-in failed.';
		}
	}
	await Browser.close().catch(() => undefined);
	if (failure) {
		await goto(`/login?error=${encodeURIComponent(failure)}`, { replaceState: true });
		return;
	}
	await loadSession();
	await goto(HOME, { replaceState: true });
}

export function registerDeepLinks(): void {
	if (registered) return;
	registered = true;
	void App.addListener('appUrlOpen', ({ url }) => {
		const link = parseDeepLink(url);
		if (link?.kind === 'auth') void finishLogin(link.code, link.state, link.error);
		else if (link?.kind === 'session') void goto(`/s/${encodeURIComponent(link.id)}`);
	});
}
