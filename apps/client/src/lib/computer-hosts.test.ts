import { describe, expect, it } from 'vitest';
import { checkHost, sameHosts } from './computer-hosts';

describe('checkHost', () => {
	it('accepts a host, a wildcard and a pinned port, normalized', () => {
		expect(checkHost(' API.Example.com ')).toEqual({ ok: true, host: 'api.example.com' });
		expect(checkHost('*.cdn.example.org')).toEqual({ ok: true, host: '*.cdn.example.org' });
		expect(checkHost('git.example.net:9418')).toEqual({ ok: true, host: 'git.example.net:9418' });
		expect(checkHost('example.com.')).toEqual({ ok: true, host: 'example.com' });
	});

	it.each([
		['', 'empty'],
		['10.0.0.1', 'ip'],
		['[::1]:443', 'ip'],
		['::1', 'ip'],
		['intranet', 'labels'],
		['*.com', 'labels'],
		['example.com:0', 'port'],
		['example.com:70000', 'port'],
		['example.com:abc', 'port'],
		['my_host.example.com', 'chars'],
		['-a.example.com', 'chars'],
		['bad host.example.com', 'chars'],
		['https://example.com', 'port'],
		['example.123', 'tld'],
		['example.0x1f', 'tld']
	])('refuses %j', (entry, problem) => {
		expect(checkHost(entry)).toEqual({ ok: false, problem });
	});
});

describe('sameHosts', () => {
	it('ignores order', () => {
		expect(sameHosts(['a.example.com', 'b.example.com'], ['b.example.com', 'a.example.com'])).toBe(true);
		expect(sameHosts(['a.example.com'], ['a.example.com', 'b.example.com'])).toBe(false);
	});
});
