import { httpApi } from './http';
import type { ChatApi } from './types';

let current: ChatApi = httpApi;

export function api(): ChatApi {
	return current;
}

export function setApi(next: ChatApi): void {
	current = next;
}

export * from './types';
