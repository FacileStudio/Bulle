import { describe, expect, it } from 'vitest';
import { BUTTON_ACTION, FORM_ACTION, toOpenUIAction } from './actions';

describe('toOpenUIAction', () => {
	it('maps a button click', () => {
		expect(
			toOpenUIAction({
				type: BUTTON_ACTION,
				params: { name: 'confirm' },
				humanFriendlyMessage: 'Confirm'
			})
		).toEqual({ type: 'button', name: 'confirm', values: {}, label: 'Confirm' });
	});

	it('flattens the submitted form values', () => {
		expect(
			toOpenUIAction({
				type: FORM_ACTION,
				params: { name: 'signup' },
				humanFriendlyMessage: 'Send',
				formName: 'signup',
				formState: {
					signup: { email: { value: 'a@b.c', componentType: 'text' }, agree: { value: true } }
				}
			})
		).toEqual({
			type: 'form',
			name: 'signup',
			values: { email: 'a@b.c', agree: true },
			label: 'Send'
		});
	});

	it('ignores events it did not emit', () => {
		expect(
			toOpenUIAction({ type: 'open_url', params: { url: 'x' }, humanFriendlyMessage: '' })
		).toBeNull();
		expect(
			toOpenUIAction({ type: BUTTON_ACTION, params: {}, humanFriendlyMessage: '' })
		).toBeNull();
	});
});
