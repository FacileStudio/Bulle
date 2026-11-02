const RELEASES = 'https://github.com/FacileStudio/Bulle/releases';

export const APK_URL = `${RELEASES}/latest/download/bulle-android.apk`;
export const APK_SHA256_URL = `${APK_URL}.sha256`;
export const RELEASES_URL = RELEASES;
export const DESKTOP_CHECKSUMS_URL = `${RELEASES}/latest/download/desktop-checksums.txt`;

export type DesktopOs = 'macos' | 'windows' | 'linux';

export interface DesktopBuild {
	os: DesktopOs;
	label: string;
	detail: string;
	url: string;
}

const latest = (asset: string) => `${RELEASES}/latest/download/${asset}`;

export const DESKTOP_BUILDS: DesktopBuild[] = [
	{ os: 'macos', label: 'macOS', detail: 'Universal .dmg', url: latest('Bulle-macos-universal.dmg') },
	{
		os: 'windows',
		label: 'Windows',
		detail: 'x64 installer',
		url: latest('Bulle-windows-x64-setup.exe')
	},
	{ os: 'linux', label: 'Linux', detail: 'x64 AppImage', url: latest('Bulle-linux-x64.AppImage') },
	{ os: 'linux', label: 'Debian, Ubuntu', detail: 'x64 .deb', url: latest('Bulle-linux-x64.deb') }
];

/** Guesses the visitor's desktop OS from a user agent, or null on phones and unknown agents. */
export function detectDesktopOs(userAgent: string): DesktopOs | null {
	if (/android|iphone|ipad|ipod/i.test(userAgent)) return null;
	if (/mac os x|macintosh/i.test(userAgent)) return 'macos';
	if (/windows/i.test(userAgent)) return 'windows';
	if (/linux|x11/i.test(userAgent)) return 'linux';
	return null;
}

export interface InstallMethod {
	label: string;
	command: string;
}

export const INSTALL_METHODS: InstallMethod[] = [
	{
		label: 'Install script',
		command: 'curl -fsSL https://raw.githubusercontent.com/FacileStudio/Bulle/main/install.sh | bash'
	},
	{ label: 'With Facile CLI', command: 'facile install bulle' },
	{ label: 'With Homebrew', command: 'brew install FacileStudio/tap/bulle' }
];
