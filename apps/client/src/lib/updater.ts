import { updater } from './updater.svelte';

export * from './updater.svelte';
export { updater, updater as updaterStore, UpdaterStore } from './updater.svelte';

export async function offerUpdate(): Promise<void> {
	await updater.checkNow(true);
}
