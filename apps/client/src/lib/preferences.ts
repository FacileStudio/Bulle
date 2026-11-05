import { apiFetch } from '$lib/backend';

export type ModelOption = { id: string; name: string };

export type Preferences = {
	model: string;
	catalog: ModelOption[];
	default_model: string;
};

export function getPreferences(): Promise<Preferences> {
	return apiFetch<Preferences>('/preferences');
}

export function updatePreferences(input: { model: string }): Promise<Preferences> {
	return apiFetch<Preferences>('/preferences', {
		method: 'PUT',
		body: JSON.stringify(input)
	});
}
