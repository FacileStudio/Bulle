import pkg from "../package.json" with { type: "json" };
import { call } from "./client.ts";
import { versionLine } from "./commands.ts";
import { serve } from "./daemon.ts";
import { paths } from "./paths.ts";
import type { Response } from "./wire.ts";

/** bulle-browser <cmd> ['<json args>'], daemon, or --version. */
async function main(argv: string[]): Promise<number> {
	const [cmd, raw] = argv;
	if (cmd === "--version") {
		process.stdout.write(`${versionLine(pkg.version)}\n`);
		return 0;
	}
	if (cmd === "daemon") {
		await serve(paths());
		return -1;
	}
	if (!cmd) {
		process.stderr.write("usage: bulle-browser <command> ['<json>'] | daemon | --version\n");
		return 2;
	}
	const response = await Promise.resolve()
		.then(() => call(paths(), cmd, raw ? JSON.parse(raw) : {}))
		.catch(
		(error: Error): Response => ({ ok: false, error: error.message }),
	);
	process.stdout.write(`${JSON.stringify(response)}\n`);
	return response.ok ? 0 : 1;
}

main(process.argv.slice(2)).then((code) => {
	if (code >= 0) {
		process.exitCode = code;
	}
});
