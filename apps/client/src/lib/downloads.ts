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

export type Platform = 'macos' | 'windows' | 'linux' | 'android';

export interface PlatformCard {
	id: Platform;
	label: string;
	summary: string;
	files: { label: string; url: string }[];
}

export const PLATFORMS: PlatformCard[] = [
	{
		id: 'macos',
		label: 'macOS',
		summary: 'Apple silicon and Intel',
		files: [{ label: 'Download .dmg', url: DESKTOP_BUILDS[0].url }]
	},
	{
		id: 'windows',
		label: 'Windows',
		summary: 'Windows 10 and 11, x64',
		files: [{ label: 'Download installer', url: DESKTOP_BUILDS[1].url }]
	},
	{
		id: 'linux',
		label: 'Linux',
		summary: 'x64 AppImage or .deb',
		files: [
			{ label: 'AppImage', url: DESKTOP_BUILDS[2].url },
			{ label: '.deb', url: DESKTOP_BUILDS[3].url }
		]
	},
	{
		id: 'android',
		label: 'Android',
		summary: 'Android 7 or later',
		files: [{ label: 'Download APK', url: APK_URL }]
	}
];

/** Picks the platform to feature: the visitor's desktop OS, Android on a phone, else macOS. */
export function featuredPlatform(userAgent: string): Platform {
	if (/android/i.test(userAgent)) return 'android';
	return detectDesktopOs(userAgent) ?? 'macos';
}

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
