import { describe, expect, test } from 'bun:test';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { cspFor, DEFAULT_API_BASE } from './csp';

describe('cspFor', () => {
	test('matches the checked-in policy for the default API', () => {
		const conf = JSON.parse(readFileSync(join(import.meta.dir, '..', 'tauri.conf.json'), 'utf8'));
		expect(cspFor(DEFAULT_API_BASE)).toBe(conf.app.security.csp);
	});

	test('lets a self-hosted or local API through', () => {
		expect(cspFor('https://bulle.example.org/api')).toContain('connect-src \'self\' ipc: http://ipc.localhost https://bulle.example.org');
		const local = cspFor('http://localhost:8080/api');
		expect(local).toContain('http://ipc.localhost http://localhost:8080');
		expect(local).toContain("img-src 'self' data: blob: https: http://localhost:8080");
	});
});
