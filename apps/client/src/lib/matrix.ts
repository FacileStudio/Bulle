import { apiFetch } from '$lib/backend';

export type MatrixRoom = {
	room_id: string;
	name: string;
	space_id: string;
};

export type MatrixStatus = {
	available: boolean;
	user_id: string;
	rooms: MatrixRoom[];
	status: string;
};

export function getMatrix(): Promise<MatrixStatus> {
	return apiFetch<MatrixStatus>('/matrix');
}

export function statusTone(status: string): 'success' | 'info' | 'danger' | 'neutral' {
	if (status === 'running') return 'success';
	if (status === 'connecting') return 'info';
	if (status.startsWith('error')) return 'danger';
	return 'neutral';
}
