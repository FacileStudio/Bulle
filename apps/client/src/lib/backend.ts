import { apiBase, clearToken, getToken, isApp } from '$lib/platform';

const backendBaseUrl = apiBase();

export const ssoLoginUrl = `${backendBaseUrl}/auth/oidc`;

export type AssistantProfile = {
	user_id?: number;
	name: string;
	avatar_url: string;
	archetype: string;
	traits: string[];
};

export type AuthUser = {
	id: number;
	email: string;
	name: string;
	created_at: string;
	assistant?: AssistantProfile | null;
};

export type PushConfig = {
	enabled: boolean;
	server: string;
	topic: string;
	subscribe_url: string;
	native: boolean;
	web_push_key?: string;
};

export type PushDevice = { kind: 'fcm' | 'web'; token: string; p256dh?: string; auth?: string };

export type MeResponse = { user: AuthUser };

export type AuthConfig = { sso_only: boolean; oidc_enabled: boolean };

type ApiErrorPayload = { error?: { message?: string } };

export class ApiError extends Error {
	status: number;

	constructor(message: string, status: number) {
		super(message);
		this.status = status;
	}
}

export async function authedInit(options: RequestInit = {}): Promise<RequestInit> {
	const headers = new Headers(options.headers);
	if (isApp()) {
		const token = await getToken();
		if (token) headers.set('Authorization', `Bearer ${token}`);
		return { ...options, headers };
	}
	if (options.method && options.method !== 'GET') {
		headers.set('X-Facile-CSRF', '1');
	}
	return { ...options, headers, credentials: 'same-origin' };
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
	const headers = new Headers(options.headers);
	if (options.body && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}

	const response = await fetch(
		`${backendBaseUrl}${path}`,
		await authedInit({ ...options, headers })
	);
	if (response.status === 401 && isApp()) await clearToken();
	if (!response.ok) {
		const payload = (await response.json().catch(() => undefined)) as ApiErrorPayload | undefined;
		throw new ApiError(
			payload?.error?.message || `Request failed with status ${response.status}`,
			response.status
		);
	}
	if (response.status === 204) return undefined as T;
	return (await response.json()) as T;
}

export function getAssistant(): Promise<AssistantProfile> {
	return apiFetch<AssistantProfile>('/assistant');
}

export function updateAssistant(profile: {
	name: string;
	avatar_url: string;
	archetype: string;
	traits: string[];
}): Promise<AssistantProfile> {
	return apiFetch<AssistantProfile>('/assistant', {
		method: 'PUT',
		body: JSON.stringify(profile)
	});
}

export const backend = {
	baseUrl: backendBaseUrl,
	authConfig() {
		return apiFetch<AuthConfig>('/auth/config');
	},
	me() {
		return apiFetch<MeResponse>('/auth/me');
	},
	async logout() {
		try {
			await apiFetch<void>('/auth/logout', { method: 'POST' });
		} finally {
			await clearToken();
		}
	},
	getPush() {
		return apiFetch<PushConfig>('/push');
	},
	testPush() {
		return apiFetch<void>('/push/test', { method: 'POST' });
	},
	registerDevice(device: PushDevice) {
		return apiFetch<void>('/push/devices', { method: 'POST', body: JSON.stringify(device) });
	},
	forgetDevice(token: string) {
		return apiFetch<void>('/push/devices', { method: 'DELETE', body: JSON.stringify({ token }) });
	},
	getAssistant,
	updateAssistant
};
