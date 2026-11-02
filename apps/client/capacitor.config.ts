import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
	appId: 'studio.facile.bulle',
	appName: 'Bulle',
	webDir: 'build',
	android: { allowMixedContent: false },
	server: { androidScheme: 'https' }
};

export default config;
