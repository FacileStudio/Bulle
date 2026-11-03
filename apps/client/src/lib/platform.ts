import { Capacitor } from '@capacitor/core';
import { isTauri } from '@tauri-apps/api/core';
import { Preferences } from '@capacitor/preferences';

export const APP_SCHEME = 'studio.facile.bulle';

const TOKEN_KEY = 'bulle.token';
const APP_API_BASE = 'https://bulle.facile.studio/api';
const VERIFIER_ALPHABET =
	'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';

export type DeepLink =
	| { kind: 'auth'; code: string; state: string; error: string }
	| { kind: 'session'; id: string };

let cachedToken: string | null | undefined;

export function isNative(): boolean {
	return Capacitor.isNativePlatform();
}

export function isDesktop(): boolean {
	return isTauri();
}

export function hasNativeTrafficLights(): boolean {
	return isDesktop() && /mac/i.test(navigator.userAgent);
}

export function isApp(): boolean {
	return isNative() || isDesktop();
}

export function apiBase(): string {
	const override = import.meta.env.VITE_API_BASE as string | undefined;
	if (override) return override;
	return isApp() ? APP_API_BASE : '/api';
}

export async function getToken(): Promise<string | null> {
	if (!isApp()) return null;
	if (cachedToken === undefined) {
		cachedToken = (await Preferences.get({ key: TOKEN_KEY })).value;
	}
	return cachedToken;
}

export async function setToken(token: string): Promise<void> {
	if (!isApp()) return;
	cachedToken = token;
	await Preferences.set({ key: TOKEN_KEY, value: token });
}

export async function clearToken(): Promise<void> {
	if (!isApp()) return;
	cachedToken = null;
	await Preferences.remove({ key: TOKEN_KEY });
}

export function base64Url(bytes: Uint8Array): string {
	let binary = '';
	for (const byte of bytes) binary += String.fromCharCode(byte);
	return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function randomString(length: number): string {
	const bytes = crypto.getRandomValues(new Uint8Array(length));
	return Array.from(bytes, (b) => VERIFIER_ALPHABET[b % VERIFIER_ALPHABET.length]).join('');
}

export function randomState(): string {
	return base64Url(crypto.getRandomValues(new Uint8Array(24)));
}

export async function pkceChallenge(verifier: string): Promise<string> {
	const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier));
	return base64Url(new Uint8Array(digest));
}

export function parseDeepLink(raw: string): DeepLink | null {
	let url: URL;
	try {
		url = new URL(raw);
	} catch {
		return null;
	}
	if (url.protocol !== `${APP_SCHEME}:`) return null;
	if (url.hostname === 'auth') {
		return {
			kind: 'auth',
			code: url.searchParams.get('code') ?? '',
			state: url.searchParams.get('state') ?? '',
			error: url.searchParams.get('error') ?? ''
		};
	}
	if (url.hostname === 'sessions') {
		const id = decodeURIComponent(url.pathname.replace(/^\/+/, '').split('/')[0] ?? '');
		return id ? { kind: 'session', id } : null;
	}
	return null;
}
