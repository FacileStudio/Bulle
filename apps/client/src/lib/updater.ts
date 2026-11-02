import { toast } from '@facile/muse';
import { relaunch } from '@tauri-apps/plugin-process';
import { check, type Update } from '@tauri-apps/plugin-updater';

async function install(update: Update): Promise<void> {
	const pending = toast.neutral(`Downloading Bulle ${update.version}…`, { duration: 0 });
	try {
		await update.downloadAndInstall();
		await relaunch();
	} catch (err) {
		toast.dismiss(pending);
		toast.danger(err instanceof Error ? err.message : 'The update could not be installed.');
	}
}

/** Checks GitHub for a newer desktop release and offers it in a pinned toast. */
export async function offerUpdate(): Promise<void> {
	const update = await check().catch(() => null);
	if (!update) return;
	toast.info(`Bulle ${update.version} is available.`, {
		duration: 0,
		action: { label: 'Update', onClick: () => void install(update) }
	});
}
