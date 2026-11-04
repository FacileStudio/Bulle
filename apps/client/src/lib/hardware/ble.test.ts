import { describe, expect, it } from 'vitest';
import { reportedHardwareId } from './ble';

describe('Pastille hardware id', () => {
	it('accepts only the 8 lowercase hex characters the device reports', () => {
		expect(reportedHardwareId({ state: 'authorized', hardware_id: '4f2a91c0' })).toBe('4f2a91c0');
		for (const bad of ['4F2A91C0', '4f2a91', '4f2a91c0aa', 'zzzzzzzz', 42, undefined]) {
			expect(reportedHardwareId({ hardware_id: bad })).toBeUndefined();
		}
		expect(reportedHardwareId(undefined)).toBeUndefined();
	});
});
