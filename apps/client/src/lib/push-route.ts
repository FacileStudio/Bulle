/** Maps a server notification path such as "sessions/abc" to an app route. */
export function routeFor(path: string | undefined): string {
	const match = /^sessions\/([^/?#]+)$/.exec(path ?? '');
	return match ? `/s/${match[1]}` : '/sessions';
}
