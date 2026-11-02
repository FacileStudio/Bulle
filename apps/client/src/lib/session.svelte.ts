import { ApiError, backend, type AssistantProfile, type AuthUser } from '$lib/backend';

type Status = 'loading' | 'signed-in' | 'signed-out' | 'unreachable';

export const session = $state<{ status: Status; user: AuthUser | null }>({
	status: 'loading',
	user: null
});

export function setAssistant(assistant: AssistantProfile): void {
	if (session.user) session.user.assistant = assistant;
}

export async function loadSession(): Promise<void> {
	try {
		session.user = (await backend.me()).user;
		session.status = 'signed-in';
	} catch (err) {
		session.user = null;
		session.status = err instanceof ApiError && err.status === 401 ? 'signed-out' : 'unreachable';
	}
}

export async function signOut(): Promise<void> {
	await backend.logout().catch(() => undefined);
	session.user = null;
	session.status = 'signed-out';
}
