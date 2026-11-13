import type { CliCommand, DownloadEntry } from '@facile/muse';
import { t } from '$lib/i18n/index.svelte';

const RELEASES = 'https://github.com/FacileStudio/Bulle/releases';

const latest = (asset: string) => `${RELEASES}/latest/download/${asset}`;

export const APK_URL = latest('bulle-android.apk');
export const APK_SHA256_URL = `${APK_URL}.sha256`;
export const RELEASES_URL = RELEASES;
export const DESKTOP_CHECKSUMS_URL = latest('desktop-checksums.txt');

export const DOWNLOADS: DownloadEntry[] = [
	{ platform: 'macos', href: latest('Bulle-macos-universal.dmg'), label: '.dmg', arch: 'universal' },
	{ platform: 'windows', href: latest('Bulle-windows-x64-setup.exe'), label: '.exe', arch: 'x64' },
	{ platform: 'linux', href: latest('Bulle-linux-x64.AppImage'), label: 'AppImage', arch: 'x64' },
	{ platform: 'linux', href: latest('Bulle-linux-x64.deb'), label: '.deb', arch: 'x64' },
	{ platform: 'android', href: APK_URL, label: 'APK', checksumHref: APK_SHA256_URL }
];

export const INSTALL_METHODS: CliCommand[] = [
	{
		get label() {
			return t('downloads.install.script');
		},
		command: 'curl -fsSL https://raw.githubusercontent.com/FacileStudio/Bulle/main/install.sh | bash'
	},
	{
		get label() {
			return t('downloads.install.facile');
		},
		command: 'facile install bulle'
	},
	{
		get label() {
			return t('downloads.install.brew');
		},
		command: 'brew install FacileStudio/tap/bulle'
	}
];
