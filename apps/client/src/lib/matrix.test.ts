import { describe, expect, it } from 'vitest';
import { statusTone } from './matrix';

describe('matrix status', () => {
	it('maps the agent status to a badge tone', () => {
		expect(statusTone('running')).toBe('success');
		expect(statusTone('connecting')).toBe('info');
		expect(statusTone('error: login failed')).toBe('danger');
		expect(statusTone('off')).toBe('neutral');
	});
});
