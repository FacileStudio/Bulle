import { describe, expect, it } from 'vitest';
import { ApiError } from './backend';

describe('ApiError', () => {
	it('reads the plan limit from a 402 envelope', () => {
		const error = new ApiError('AI features need a paid plan.', 402, {
			error: {
				code: 'payment_required',
				message: 'AI features need a paid plan.',
				upgrade_url: 'https://sso.facile.studio/spaces'
			}
		});
		expect(error.planLimit?.upgrade_url).toBe('https://sso.facile.studio/spaces');
		expect(error.planLimit?.message).toBe('AI features need a paid plan.');
	});

	it('drops an unsafe upgrade link', () => {
		const error = new ApiError('x', 402, { error: { message: 'x', upgrade_url: 'javascript:alert(1)' } });
		expect(error.planLimit?.upgrade_url).toBeUndefined();
	});

	it('leaves other statuses without a plan limit', () => {
		expect(new ApiError('boom', 500, { error: { message: 'boom' } }).planLimit).toBeNull();
	});
});
