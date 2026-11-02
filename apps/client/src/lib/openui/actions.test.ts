import { describe, expect, it } from 'vitest';
import { BUTTON_ACTION, FORM_ACTION, actionMessage, toOpenUIAction } from './actions';

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

describe('actionMessage', () => {
	it('formats button actions with label', () => {
		expect(
			actionMessage({ type: 'button', name: 'retry', values: {}, label: 'Retry' })
		).toBe('Retry');
	});

	it('formats form actions with json values', () => {
		expect(
			actionMessage({
				type: 'form',
				name: 'survey',
				values: { score: 5 },
				label: 'Submit Survey'
			})
		).toBe('Submit Survey\n{"score":5}');
	});
});
