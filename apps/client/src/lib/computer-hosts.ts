export const MAX_HOSTS = 64;

export type HostProblem = 'empty' | 'ip' | 'port' | 'labels' | 'chars' | 'tld';

export type HostCheck = { ok: true; host: string } | { ok: false; problem: HostProblem };

const LABEL = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/;
const IPV4 = /^\d{1,3}(\.\d{1,3}){3}$/;
const HEX_NUMBER = /^0x[0-9a-f]*$/;

function bad(problem: HostProblem): HostCheck {
	return { ok: false, problem };
}

/**
 * Checks one allowed host with the worker's own rules: `host`, `*.host`
 * (subdomains only) or `host:port`. IP literals, single-label names, bare
 * wildcards such as `*.com` and bad ports are refused. It lowercases and drops
 * one trailing dot, as the server does.
 */
export function checkHost(entry: string): HostCheck {
	let rest = entry.trim().toLowerCase();
	if (!rest) return bad('empty');
	if (rest.startsWith('[') || rest.split(':').length > 2) return bad('ip');
	let port = '';
	const colon = rest.indexOf(':');
	if (colon >= 0) {
		const raw = rest.slice(colon + 1);
		const n = Number(raw);
		if (!/^\d+$/.test(raw) || n < 1 || n > 65535) return bad('port');
		port = String(n);
		rest = rest.slice(0, colon);
	}
	let wildcard = false;
	if (rest.startsWith('*.')) {
		wildcard = true;
		rest = rest.slice(2);
	}
	if (rest.endsWith('.')) rest = rest.slice(0, -1);
	if (!rest || rest.length > 253) return bad('labels');
	if (IPV4.test(rest)) return bad('ip');
	const labels = rest.split('.');
	if (labels.length < 2) return bad('labels');
	if (!labels.every((label) => LABEL.test(label))) return bad('chars');
	const tld = labels[labels.length - 1];
	if (!/[a-z]/.test(tld) || HEX_NUMBER.test(tld)) return bad('tld');
	return { ok: true, host: `${wildcard ? '*.' : ''}${rest}${port ? `:${port}` : ''}` };
}

/** The same hosts in the same set, whatever their order. */
export function sameHosts(a: readonly string[], b: readonly string[]): boolean {
	if (a.length !== b.length) return false;
	const sorted = [...b].sort();
	return [...a].sort().every((host, i) => host === sorted[i]);
}
