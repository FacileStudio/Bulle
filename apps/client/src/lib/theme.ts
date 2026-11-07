import { applyTheme, isTheme, type Theme } from '@facile/muse';

export const THEME_KEY = 'bulle-theme';

export function storedTheme(): Theme {
	try {
		const value = localStorage.getItem(THEME_KEY);
		return isTheme(value) ? value : 'system';
	} catch {
		return 'system';
	}
}

export function chooseTheme(theme: Theme) {
	applyTheme(theme);
	try {
		localStorage.setItem(THEME_KEY, theme);
	} catch {
		return;
	}
}
