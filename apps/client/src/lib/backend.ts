const backendBaseUrl = '/api';

export const ssoLoginUrl = `${backendBaseUrl}/auth/oidc`;

export type AuthUser = {
	id: number;
	email: string;
	name: string;
	created_at: string;
};

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

async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
	const headers = new Headers(options.headers);
	if (options.body && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}
	if (options.method && options.method !== 'GET') {
		headers.set('X-Facile-CSRF', '1');
	}

	const response = await fetch(`${backendBaseUrl}${path}`, {
		...options,
		headers,
		credentials: 'same-origin'
	});
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

export const backend = {
	baseUrl: backendBaseUrl,
	authConfig() {
		return apiFetch<AuthConfig>('/auth/config');
	},
	me() {
		return apiFetch<MeResponse>('/auth/me');
	},
	logout() {
		return apiFetch<void>('/auth/logout', { method: 'POST' });
	}
};
