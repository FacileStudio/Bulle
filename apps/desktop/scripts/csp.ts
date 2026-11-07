import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname } from 'node:path';

export const DEFAULT_API_BASE = 'https://bulle.facile.studio/api';

/**
 * Builds the desktop content security policy for the API the client talks to.
 * The API origin joins connect-src, and img-src and media-src when it is not
 * already covered by https:.
 */
export function cspFor(apiBase: string): string {
	const origin = new URL(apiBase).origin;
	const media = origin.startsWith('https:') ? '' : ` ${origin}`;
	return [
		"default-src 'self'",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline'",
		`img-src 'self' data: blob: https:${media}`,
		"font-src 'self' data:",
		`media-src 'self' data: blob: https:${media}`,
		`connect-src 'self' ipc: http://ipc.localhost ${origin}`
	].join('; ');
}

/** Writes the Tauri config override that `tauri build --config` merges, for VITE_API_BASE or the default. */
function main(): void {
	const out = process.argv[2];
	if (!out) throw new Error('Pass the output path, for example gen/csp.json');
	const apiBase = process.env.VITE_API_BASE || DEFAULT_API_BASE;
	mkdirSync(dirname(out), { recursive: true });
	writeFileSync(out, `${JSON.stringify({ app: { security: { csp: cspFor(apiBase) } } }, null, 2)}\n`);
	console.log(`Wrote ${out} for ${apiBase}`);
}

if (import.meta.main) main();
