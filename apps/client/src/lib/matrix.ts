import { apiFetch } from '$lib/backend';

export type MatrixSettings = {
	enabled: boolean;
	homeserver: string;
	user_id: string;
	password?: string;
	password_set: boolean;
	allow: string[];
	rooms: Record<string, string>;
	status: string;
};

export function getMatrix(): Promise<MatrixSettings> {
	return apiFetch<MatrixSettings>('/matrix');
}

export function saveMatrix(settings: MatrixSettings): Promise<MatrixSettings> {
	return apiFetch<MatrixSettings>('/matrix', { method: 'PUT', body: JSON.stringify(settings) });
}

export function parseAllow(text: string): string[] {
	return text
		.split(/[\s,]+/)
		.map((s) => s.trim())
		.filter(Boolean);
}

export function statusTone(status: string): 'success' | 'info' | 'danger' | 'neutral' {
	if (status === 'running') return 'success';
	if (status === 'connecting') return 'info';
	if (status.startsWith('error')) return 'danger';
	return 'neutral';
}
