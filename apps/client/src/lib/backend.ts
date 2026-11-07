import { planLimitFrom, type PlanLimitError } from '@facile/muse';
import { t } from '$lib/i18n/index.svelte';
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
	avatar_url?: string;
	locale?: string;
	plan?: 'free' | 'paid';
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
	planLimit: PlanLimitError | null;

	constructor(message: string, status: number, payload?: unknown) {
		super(message);
		this.status = status;
		this.planLimit = status === 402 ? planLimitFrom(payload) : null;
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
			payload?.error?.message || t('error.requestFailed', { status: response.status }),
			response.status,
			payload
		);
	}
	if (response.status === 204) return undefined as T;
	return (await response.json()) as T;
}

function errorFrom(status: number, raw: string): ApiError {
	let payload: ApiErrorPayload | undefined;
	try {
		payload = JSON.parse(raw) as ApiErrorPayload;
	} catch {
		payload = undefined;
	}
	return new ApiError(payload?.error?.message || t('error.requestFailed', { status }), status, payload);
}

export function uploadFile<T>(
	path: string,
	body: FormData,
	onProgress?: (fraction: number) => void
): { done: Promise<T>; abort: () => void } {
	const xhr = new XMLHttpRequest();
	let aborted = false;
	const done = (async () => {
		const init = await authedInit({ method: 'POST' });
		return new Promise<T>((resolve, reject) => {
			if (aborted) {
				reject(new ApiError(t('error.uploadCancelled'), 0));
				return;
			}
			xhr.open('POST', `${backendBaseUrl}${path}`);
			new Headers(init.headers).forEach((value, key) => xhr.setRequestHeader(key, value));
			xhr.upload.onprogress = (event) => {
				if (event.lengthComputable) onProgress?.(event.loaded / event.total);
			};
			xhr.onload = () => {
				if (xhr.status === 401 && isApp()) void clearToken();
				if (xhr.status < 200 || xhr.status >= 300) {
					reject(errorFrom(xhr.status, xhr.responseText));
					return;
				}
				try {
					resolve(JSON.parse(xhr.responseText) as T);
				} catch {
					reject(new ApiError(t('error.unreadable'), xhr.status));
				}
			};
			xhr.onerror = () => reject(new ApiError(t('error.uploadUnreachable'), 0));
			xhr.onabort = () => reject(new ApiError(t('error.uploadCancelled'), 0));
			xhr.send(body);
		});
	})();
	return {
		done,
		abort: () => {
			aborted = true;
			xhr.abort();
		}
	};
}

export function attachmentUrl(id: string): string {
	return `${backendBaseUrl}/attachments/${encodeURIComponent(id)}`;
}

export async function attachmentObjectUrl(id: string): Promise<string> {
	const response = await fetch(attachmentUrl(id), await authedInit());
	if (response.status === 401 && isApp()) await clearToken();
	if (!response.ok) throw errorFrom(response.status, await response.text().catch(() => ''));
	return URL.createObjectURL(await response.blob());
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

export type UsageTotals = {
	runs: number;
	input_tokens: number;
	output_tokens: number;
	cache_read_tokens: number;
	cache_creation_tokens: number;
	cost_usd: number;
};

export type UsageGroup = UsageTotals & { name: string };

export type UsageSummary = {
	timezone: string;
	today: UsageTotals;
	week: UsageTotals;
	month: UsageTotals;
	days: (UsageTotals & { date: string })[];
	models: UsageGroup[];
	channels: UsageGroup[];
};

export function getUsage(): Promise<UsageSummary> {
	const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
	return apiFetch<UsageSummary>(`/usage?tz=${encodeURIComponent(tz)}`);
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
