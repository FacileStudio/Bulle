import { browser } from '$app/environment';
import { createTranslator, LOCALE_COOKIE, resolveLocale, type Locale } from '@facile/muse';
import { catalogs } from './messages';

let active: () => Locale = () => 'en';

export function localeSource(sessionLocale: () => string | null | undefined): string | null {
	return sessionLocale() ?? cookieLocale() ?? (browser ? navigator.language : null);
}

function cookieLocale(): string | null {
	if (!browser) return null;
	const prefix = `${LOCALE_COOKIE}=`;
	const found = document.cookie.split('; ').find((part) => part.startsWith(prefix));
	return found ? decodeURIComponent(found.slice(prefix.length)) : null;
}

export function rememberLocale(locale: Locale) {
	document.cookie = `${LOCALE_COOKIE}=${locale}; path=/; max-age=31536000; samesite=lax`;
}

export function bindLocale(source: () => Locale) {
	active = source;
}

export function locale(): Locale {
	return resolveLocale(active());
}

export const t = createTranslator(catalogs, () => active());
