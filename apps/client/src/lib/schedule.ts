import { t } from '$lib/i18n/index.svelte';

const DAYS: Record<string, () => string> = {
	'*': () => t('jobs.schedule.everyDay'),
	'1-5': () => t('jobs.schedule.weekdays'),
	'0,6': () => t('jobs.schedule.weekends'),
	'6,0': () => t('jobs.schedule.weekends')
};

export function describeSchedule(expr: string): string {
	const spec = expr.replace(/^(CRON_)?TZ=\S+\s+/, '').trim();
	const parts = spec.split(/\s+/);
	if (parts.length !== 5) return spec;
	const [minute, hour, dom, month, dow] = parts;
	const days = DAYS[dow]?.();
	if (!/^\d+$/.test(minute) || !/^\d+$/.test(hour) || dom !== '*' || month !== '*' || !days) return spec;
	return t('jobs.schedule.at', { days, time: `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}` });
}
