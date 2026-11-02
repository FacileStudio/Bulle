import { describe, expect, it } from 'vitest';
import { describeSchedule } from './schedule';

describe('describeSchedule', () => {
	it('reads common daily schedules and keeps the rest raw', () => {
		expect(describeSchedule('CRON_TZ=Europe/Paris 0 8 * * 1-5')).toBe('Weekdays at 08:00');
		expect(describeSchedule('30 19 * * *')).toBe('Every day at 19:30');
		expect(describeSchedule('*/15 * * * *')).toBe('*/15 * * * *');
		expect(describeSchedule('@daily')).toBe('@daily');
	});
});
