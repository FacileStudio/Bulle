import { App } from '@capacitor/app';
import { Browser } from '@capacitor/browser';
import { isTauri } from '@tauri-apps/api/core';
import { check, type DownloadEvent, type Update } from '@tauri-apps/plugin-updater';
import { relaunch } from '@tauri-apps/plugin-process';
import { APK_URL } from '$lib/downloads';
import { t } from '$lib/i18n/index.svelte';
import { apiBase, isNative } from '$lib/platform';
import pkg from '../../package.json';

const LATEST_RELEASE = 'https://api.github.com/repos/FacileStudio/Bulle/releases/latest';

export type UpdaterState =
	| 'idle'
	| 'checking'
	| 'available'
	| 'downloading'
	| 'ready'
	| 'error'
	| 'upToDate';

function parts(version: string): number[] {
	return version.replace(/^v/, '').split(/[.-]/).slice(0, 3).map((p) => Number.parseInt(p, 10) || 0);
}

/** Compares two semantic versions such as "3.27.0" and "v3.28.1": negative, zero or positive. */
export function compareVersions(a: string, b: string): number {
	const [x, y] = [parts(a), parts(b)];
	for (let i = 0; i < 3; i++) if ((x[i] ?? 0) !== (y[i] ?? 0)) return (x[i] ?? 0) - (y[i] ?? 0);
	return 0;
}

async function fetchVersion(url: string, field: 'version' | 'tag_name'): Promise<string> {
	const response = await fetch(url, { headers: { Accept: 'application/json' } });
	if (!response.ok) throw new Error(t('error.requestFailed', { status: response.status }));
	const body = (await response.json()) as Record<string, unknown>;
	const value = body[field];
	if (typeof value !== 'string' || !value) throw new Error(t('error.unreadable'));
	return value.replace(/^v/, '');
}

export class UpdaterStore {
	state = $state<UpdaterState>('idle');
	currentVersion = $state<string>(pkg.version);
	availableVersion = $state<string | null>(null);
	downloadProgress = $state<number>(0);
	errorMessage = $state<string | null>(null);

	private updateHandle: Update | null = null;
	private downloadedBytes = 0;
	private totalBytes = 0;

	checkNow = async (auto = false): Promise<void> => {
		if (this.state === 'checking' || this.state === 'downloading') return;
		this.errorMessage = null;
		if (!auto) {
			this.state = 'checking';
		}

		if (isNative()) {
			await this.checkAndroid(auto);
			return;
		}

		if (!isTauri()) {
			this.state = 'idle';
			return;
		}

		try {
			const update = await check();
			if (update) {
				this.updateHandle = update;
				this.currentVersion = update.currentVersion || this.currentVersion;
				this.availableVersion = update.version;
				this.state = 'available';
			} else {
				this.updateHandle = null;
				this.availableVersion = null;
				this.state = 'upToDate';
			}
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (!auto) {
				this.errorMessage = msg;
				this.state = 'error';
			} else {
				this.state = 'idle';
			}
		}
	};

	/**
	 * Compares the installed APK with the server's /version and the latest GitHub
	 * release. It says up to date only when the GitHub check answered.
	 */
	private async checkAndroid(auto: boolean): Promise<void> {
		const installed = await App.getInfo()
			.then((info) => info.version)
			.catch(() => this.currentVersion);
		this.currentVersion = installed;
		const [server, release] = await Promise.allSettled([
			fetchVersion(`${apiBase()}/version`, 'version'),
			fetchVersion(LATEST_RELEASE, 'tag_name')
		]);
		const newer = [server, release]
			.flatMap((r) => (r.status === 'fulfilled' ? [r.value] : []))
			.filter((v) => compareVersions(v, installed) > 0)
			.sort(compareVersions)
			.at(-1);
		if (newer) {
			this.availableVersion = newer;
			this.state = 'available';
			return;
		}
		this.availableVersion = null;
		if (release.status === 'fulfilled') {
			this.state = 'upToDate';
			return;
		}
		const reason = release.reason;
		this.errorMessage = auto ? null : reason instanceof Error ? reason.message : String(reason);
		this.state = auto ? 'idle' : 'error';
	}

	installNow = async (): Promise<void> => {
		if (this.state === 'downloading' || this.state === 'ready') return;

		if (isNative()) {
			await Browser.open({ url: APK_URL });
			return;
		}

		if (!isTauri()) return;

		if (!this.updateHandle) {
			this.errorMessage = t('updater.noUpdate');
			this.state = 'error';
			return;
		}

		this.state = 'downloading';
		this.downloadProgress = 0;
		this.downloadedBytes = 0;
		this.totalBytes = 0;

		try {
			await this.updateHandle.downloadAndInstall((event: DownloadEvent) => {
				if (event.event === 'Started') {
					this.totalBytes = event.data.contentLength ?? 0;
					this.downloadedBytes = 0;
					this.downloadProgress = 0;
				} else if (event.event === 'Progress') {
					this.downloadedBytes += event.data.chunkLength;
					if (this.totalBytes > 0) {
						this.downloadProgress = Math.min(
							99,
							Math.round((this.downloadedBytes / this.totalBytes) * 100)
						);
					}
				} else if (event.event === 'Finished') {
					this.downloadProgress = 100;
				}
			});
			this.downloadProgress = 100;
			this.state = 'ready';
		} catch (err) {
			this.errorMessage = err instanceof Error ? err.message : String(err);
			this.state = 'error';
		}
	};

	restartNow = async (): Promise<void> => {
		if (!isTauri()) {
			if (typeof window !== 'undefined') {
				window.location.reload();
			}
			return;
		}
		try {
			await relaunch();
		} catch (err) {
			this.errorMessage = err instanceof Error ? err.message : String(err);
			this.state = 'error';
		}
	};
}

export const updaterStore = new UpdaterStore();
export const updater = updaterStore;
