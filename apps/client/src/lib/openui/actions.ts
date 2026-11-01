import type { ActionEvent } from '@openuidev/lang-core';

export const BUTTON_ACTION = 'bulle_button';
export const FORM_ACTION = 'bulle_form';

export type OpenUIAction = {
	type: 'button' | 'form';
	name: string;
	values: Record<string, unknown>;
	label: string;
};

type FormState = Record<string, { value?: unknown } | undefined>;

function flatten(formName: string | undefined, formState: ActionEvent['formState']) {
	const entries = formName ? (formState?.[formName] as FormState | undefined) : undefined;
	const values: Record<string, unknown> = {};
	for (const [key, field] of Object.entries(entries ?? {})) {
		if (field && field.value !== undefined) values[key] = field.value;
	}
	return values;
}

export function toOpenUIAction(event: ActionEvent): OpenUIAction | null {
	const name = event.params?.name;
	if (typeof name !== 'string' || name === '') return null;
	if (event.type === FORM_ACTION) {
		return {
			type: 'form',
			name,
			values: flatten(event.formName ?? name, event.formState),
			label: event.humanFriendlyMessage
		};
	}
	if (event.type === BUTTON_ACTION) {
		return { type: 'button', name, values: {}, label: event.humanFriendlyMessage };
	}
	return null;
}

export function actionMessage(action: OpenUIAction): string {
	if (action.type === 'button' || Object.keys(action.values).length === 0) return action.label;
	return `${action.label}\n${JSON.stringify(action.values)}`;
}
