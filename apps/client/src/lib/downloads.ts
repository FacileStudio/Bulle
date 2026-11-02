const RELEASES = 'https://github.com/FacileStudio/Bulle/releases';

export const APK_URL = `${RELEASES}/latest/download/bulle-android.apk`;
export const APK_SHA256_URL = `${APK_URL}.sha256`;
export const RELEASES_URL = RELEASES;

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
