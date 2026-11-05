import { apiFetch } from '$lib/backend';

export type Preferences = {
	model: string;
	cli_config: string;
	models: string[];
	default_model: string;
};

export type ModelOffer = {
	backend: string;
	model: string;
	models: string[];
};

export function getPreferences(): Promise<Preferences> {
	return apiFetch<Preferences>('/preferences');
}

export function updatePreferences(input: {
	model?: string;
	cli_config?: string;
}): Promise<Preferences> {
	return apiFetch<Preferences>('/preferences', {
		method: 'PUT',
		body: JSON.stringify(input)
	});
}

export function getModelOffer(): Promise<ModelOffer> {
	return apiFetch<ModelOffer>('/llm');
}
