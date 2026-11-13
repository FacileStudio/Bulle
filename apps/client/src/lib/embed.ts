import type { PageContext } from '$lib/api';

export const EMBED_APPS = ['docs', 'toile', 'feuilles'] as const;

export type EmbedApp = (typeof EMBED_APPS)[number];

export type EmbedAsk = { text: string; context: PageContext | null };

const SUITE_ORIGINS = EMBED_APPS.map((app) => `https://${app}.facile.studio`);
const MAX_FIELD = 300;
const MAX_SELECTION = 8000;

/**
 * Reports whether a page at this origin may talk to the embedded chat: the
 * suite apps that frame it, and in development any local port.
 */
export function isAllowedOrigin(origin: string, dev: boolean): boolean {
	if (SUITE_ORIGINS.includes(origin)) return true;
	if (!dev) return false;
	try {
		const url = new URL(origin);
		return url.protocol === 'http:' && (url.hostname === 'localhost' || url.hostname === '127.0.0.1');
	} catch {
		return false;
	}
}

/**
 * The origin of the page framing this one, or null when it is not one the
 * chat answers. Chromium and Safari name it in ancestorOrigins; Firefox only
 * through the referrer.
 */
export function parentOrigin(
	location: { ancestorOrigins?: { length: number; [index: number]: string } },
	referrer: string,
	dev: boolean
): string | null {
	const candidates = [location.ancestorOrigins?.length ? location.ancestorOrigins[0] : '', originOf(referrer)];
	return candidates.find((origin) => origin && isAllowedOrigin(origin, dev)) ?? null;
}

function originOf(url: string): string {
	try {
		return new URL(url).origin;
	} catch {
		return '';
	}
}

function line(value: unknown): string {
	return typeof value === 'string' ? value.replace(/\s+/g, ' ').trim().slice(0, MAX_FIELD) : '';
}

/** Reads the open page from the embed URL or a message, or null when it names none. */
export function pageFrom(field: (name: string) => unknown): PageContext | null {
	const app = line(field('app')).toLowerCase();
	const page = { app, id: line(field('id')), title: line(field('title')), url: line(field('url')) };
	if (!EMBED_APPS.includes(app as EmbedApp) || !page.id || !/^https?:\/\//.test(page.url)) return null;
	return page;
}

/** Reads a bulle:ask message, or null when the data is anything else. */
export function askFrom(data: unknown): EmbedAsk | null {
	if (!data || typeof data !== 'object') return null;
	const message = data as Record<string, unknown>;
	if (message.type !== 'bulle:ask' || typeof message.text !== 'string' || !message.text.trim()) return null;
	const context = message.context && typeof message.context === 'object' ? (message.context as Record<string, unknown>) : null;
	return { text: message.text.trim().slice(0, MAX_SELECTION), context: context ? pageFrom((name) => context[name]) : null };
}

/** The message that sends a selection: the lead line, then the selection as a quote. */
export function quoteSelection(lead: string, text: string): string {
	const quoted = text
		.trim()
		.split('\n')
		.map((row) => `> ${row}`.trimEnd())
		.join('\n');
	return `${lead}\n\n${quoted}`;
}

/** Where the conversation for one page is remembered, per user. */
export function conversationKey(userId: number, page: PageContext): string {
	return `bulle:embed:${userId}:${page.app}:${page.id}`;
}

/** Whether two pages are the same page with the same title and address. */
export function samePage(a: PageContext | null, b: PageContext | null): boolean {
	return !!a && !!b && a.app === b.app && a.id === b.id && a.title === b.title && a.url === b.url;
}
