import { describe, expect, it } from 'vitest';
import { askFrom, conversationKey, isAllowedOrigin, pageFrom, parentOrigin, quoteSelection, samePage } from './embed';

describe('isAllowedOrigin', () => {
	it('answers only the suite apps that frame the chat', () => {
		expect(isAllowedOrigin('https://docs.facile.studio', false)).toBe(true);
		expect(isAllowedOrigin('https://toile.facile.studio', false)).toBe(true);
		expect(isAllowedOrigin('https://feuilles.facile.studio', false)).toBe(true);
		expect(isAllowedOrigin('https://courrier.facile.studio', false)).toBe(false);
		expect(isAllowedOrigin('https://docs.facile.studio.evil.test', false)).toBe(false);
		expect(isAllowedOrigin('http://docs.facile.studio', false)).toBe(false);
		expect(isAllowedOrigin('null', false)).toBe(false);
	});

	it('lets a local port in only in development', () => {
		expect(isAllowedOrigin('http://localhost:5173', false)).toBe(false);
		expect(isAllowedOrigin('http://localhost:5173', true)).toBe(true);
		expect(isAllowedOrigin('http://127.0.0.1:3000', true)).toBe(true);
		expect(isAllowedOrigin('http://localhost.evil.test', true)).toBe(false);
	});
});

describe('parentOrigin', () => {
	it('prefers ancestorOrigins and falls back to the referrer', () => {
		expect(parentOrigin({ ancestorOrigins: ['https://toile.facile.studio'] }, '', false)).toBe('https://toile.facile.studio');
		expect(parentOrigin({}, 'https://docs.facile.studio/p/1', false)).toBe('https://docs.facile.studio');
		expect(parentOrigin({ ancestorOrigins: ['https://evil.test'] }, 'https://evil.test/', false)).toBeNull();
	});
});

describe('pageFrom', () => {
	const params = (query: string) => {
		const search = new URLSearchParams(query);
		return (name: string) => search.get(name);
	};

	it('reads the open page from the embed URL', () => {
		expect(pageFrom(params('app=docs&id=p1&title=Plan%20Q4&url=https%3A%2F%2Fdocs.facile.studio%2Fp%2Fp1'))).toEqual({
			app: 'docs',
			id: 'p1',
			title: 'Plan Q4',
			url: 'https://docs.facile.studio/p/p1'
		});
	});

	it('refuses another app, a missing id or a script URL', () => {
		expect(pageFrom(params('app=opus&id=1&url=https://opus.facile.studio'))).toBeNull();
		expect(pageFrom(params('app=docs&url=https://docs.facile.studio'))).toBeNull();
		expect(pageFrom(params('app=docs&id=1&url=javascript:alert(1)'))).toBeNull();
	});
});

describe('askFrom', () => {
	it('reads a bulle:ask with its context', () => {
		const ask = askFrom({ type: 'bulle:ask', text: ' the line ', context: { app: 'feuilles', id: 'w1', title: 'Budget', url: 'https://feuilles.facile.studio/w/w1' } });
		expect(ask).toEqual({ text: 'the line', context: { app: 'feuilles', id: 'w1', title: 'Budget', url: 'https://feuilles.facile.studio/w/w1' } });
	});

	it('ignores every other message', () => {
		expect(askFrom({ type: 'bulle:close' })).toBeNull();
		expect(askFrom({ type: 'bulle:ask', text: '  ' })).toBeNull();
		expect(askFrom('bulle:ask')).toBeNull();
		expect(askFrom({ type: 'bulle:ask', text: 'x', context: { app: 'opus' } })).toEqual({ text: 'x', context: null });
	});
});

describe('helpers', () => {
	it('quotes a selection under a lead line', () => {
		expect(quoteSelection('About this passage:', 'one\n\ntwo')).toBe('About this passage:\n\n> one\n>\n> two');
	});

	it('keys a conversation per user and page', () => {
		const page = { app: 'docs', id: 'p1', title: 'T', url: 'https://docs.facile.studio/p/p1' };
		expect(conversationKey(7, page)).toBe('bulle:embed:7:docs:p1');
		expect(samePage(page, { ...page })).toBe(true);
		expect(samePage(page, { ...page, title: 'U' })).toBe(false);
		expect(samePage(null, page)).toBe(false);
	});
});
