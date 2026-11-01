import { writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { createServer } from 'vite';

const root = fileURLToPath(new URL('../../..', import.meta.url));
const target = fileURLToPath(new URL('./prompt.generated.txt', import.meta.url));

const server = await createServer({
	root,
	logLevel: 'error',
	server: { middlewareMode: true, hmr: false, watch: null }
});

try {
	const mod = await server.ssrLoadModule('/src/lib/openui/prompt-source.ts');
	writeFileSync(target, `${mod.buildPrompt()}\n`);
	console.log(`wrote ${target}`);
} finally {
	await server.close();
}
