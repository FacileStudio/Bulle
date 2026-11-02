const DAYS: Record<string, string> = {
	'*': 'Every day',
	'1-5': 'Weekdays',
	'0,6': 'Weekends',
	'6,0': 'Weekends'
};

export function describeSchedule(expr: string): string {
	const spec = expr.replace(/^(CRON_)?TZ=\S+\s+/, '').trim();
	const parts = spec.split(/\s+/);
	if (parts.length !== 5) return spec;
	const [minute, hour, dom, month, dow] = parts;
	const days = DAYS[dow];
	if (!/^\d+$/.test(minute) || !/^\d+$/.test(hour) || dom !== '*' || month !== '*' || !days) return spec;
	return `${days} at ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`;
}
