export const HOME = '/sessions';

export function safeNext(raw: string | null): string {
	if (!raw || !raw.startsWith('/')) return HOME;
	if (raw.startsWith('//') || raw.startsWith('/\\')) return HOME;
	if (raw === '/login' || raw.startsWith('/login/') || raw.startsWith('/login?')) return HOME;
	return raw;
}
