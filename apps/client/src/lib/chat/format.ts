export function formatDuration(ms: number | undefined): string {
	if (ms === undefined) return '';
	if (ms < 1000) return `${Math.round(ms)} ms`;
	const seconds = ms / 1000;
	if (seconds < 60) return `${seconds.toFixed(1)} s`;
	return `${Math.floor(seconds / 60)} min ${Math.round(seconds % 60)} s`;
}

export function formatTokens(tokens: number): string {
	if (tokens < 1000) return String(tokens);
	if (tokens < 1_000_000) return `${(tokens / 1000).toFixed(1)}k`;
	return `${(tokens / 1_000_000).toFixed(2)}M`;
}

export function formatCost(usd: number): string {
	if (usd === 0) return '$0.00';
	if (usd < 0.01) return '<$0.01';
	return `$${usd.toFixed(2)}`;
}

export function relativeTime(iso: string, now = Date.now()): string {
	const then = Date.parse(iso);
	if (Number.isNaN(then)) return '';
	const minutes = Math.max(0, Math.round((now - then) / 60000));
	if (minutes < 1) return 'now';
	if (minutes < 60) return `${minutes} min`;
	const hours = Math.round(minutes / 60);
	if (hours < 24) return `${hours} h`;
	return `${Math.round(hours / 24)} d`;
}
