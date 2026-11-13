import { Capacitor } from '@capacitor/core';
import { invoke, isTauri } from '@tauri-apps/api/core';
import { Preferences } from '@capacitor/preferences';
import { SecureStoragePlugin } from 'capacitor-secure-storage-plugin';

export const APP_SCHEME = 'studio.facile.bulle';

const TOKEN_KEY = 'bulle.token';
const APP_API_BASE = 'https://bulle.facile.studio/api';
const VERIFIER_ALPHABET =
	'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';

export type DeepLink =
	| { kind: 'auth'; code: string; state: string; error: string }
	| { kind: 'session'; id: string };

let cachedToken: string | null | undefined;
let loadingToken: Promise<string | null> | undefined;

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

/** The token in the secure store (Keystore, Keychain, Credential Manager, Secret Service), or null. */
async function readSecureToken(): Promise<string | null> {
	try {
		if (isDesktop()) return (await invoke<string | null>('token_get')) ?? null;
		return (await SecureStoragePlugin.get({ key: TOKEN_KEY })).value || null;
	} catch {
		return null;
	}
}

/** Writes `token` to the secure store and reports whether the store kept it. */
async function writeSecureToken(token: string): Promise<boolean> {
	try {
		if (isDesktop()) await invoke('token_set', { token });
		else await SecureStoragePlugin.set({ key: TOKEN_KEY, value: token });
		return true;
	} catch (error) {
		console.warn('bulle: the secure store refused the token', error);
		return false;
	}
}

/**
 * Moves the token older builds kept in Preferences (SharedPreferences on Android, localStorage on
 * desktop) into the secure store, so the update signs nobody out. The old copy goes only once the
 * secure store holds the token.
 */
async function migratePreferencesToken(): Promise<string | null> {
	try {
		const { value } = await Preferences.get({ key: TOKEN_KEY });
		if (!value) return null;
		if (await writeSecureToken(value)) await Preferences.remove({ key: TOKEN_KEY });
		return value;
	} catch {
		return null;
	}
}

async function loadToken(): Promise<string | null> {
	return (await readSecureToken()) ?? (await migratePreferencesToken());
}

/** The installed app's bearer token, or null on the web and when signed out. */
export async function getToken(): Promise<string | null> {
	if (!isApp()) return null;
	if (cachedToken !== undefined) return cachedToken;
	loadingToken ??= loadToken();
	const token = await loadingToken;
	if (cachedToken === undefined) cachedToken = token;
	return cachedToken;
}

/** Keeps `token` in the secure store. It stays in memory for this run even if the store refuses it. */
export async function setToken(token: string): Promise<void> {
	if (!isApp()) return;
	cachedToken = token;
	await writeSecureToken(token);
}

async function clearSecureToken(): Promise<void> {
	if (isDesktop()) await invoke('token_clear');
	else await SecureStoragePlugin.remove({ key: TOKEN_KEY });
}

/** Forgets the bearer token, signing the installed app out, along with any copy left in Preferences. */
export async function clearToken(): Promise<void> {
	if (!isApp()) return;
	cachedToken = null;
	await clearSecureToken().catch(() => undefined);
	await Preferences.remove({ key: TOKEN_KEY }).catch(() => undefined);
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
