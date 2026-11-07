import { locale } from './index.svelte';

export function formatDate(value: Date | string | number, options: Intl.DateTimeFormatOptions = {}): string {
	return new Intl.DateTimeFormat(locale(), options).format(new Date(value));
}

export function formatNumber(value: number, options: Intl.NumberFormatOptions = {}): string {
	return new Intl.NumberFormat(locale(), options).format(value);
}

export function formatRelative(value: number, unit: Intl.RelativeTimeFormatUnit): string {
	return new Intl.RelativeTimeFormat(locale(), { numeric: 'auto' }).format(value, unit);
}
