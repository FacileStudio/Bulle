import { spawn } from "node:child_process";
import { mkdirSync, openSync } from "node:fs";

import type { Args } from "./commands.ts";
import type { Paths } from "./paths.ts";
import { reachable, send, type Response } from "./wire.ts";

const PASSIVE = new Set(["abort", "cancel", "status"]);
const START_TIMEOUT_MS = 60_000;

/**
 * Sends one command to the daemon, starting it first when it is not running.
 * A command that only stops things never starts a browser to do it.
 */
export async function call(where: Paths, cmd: string, args: Args): Promise<Response> {
	try {
		return await send(where.socket, { cmd, args });
	} catch (error) {
		if (!notRunning(error)) {
			throw error;
		}
	}
	if (PASSIVE.has(cmd)) {
		return { ok: true, text: "The browser is not running.", held: [] };
	}
	await start(where);
	return send(where.socket, { cmd, args });
}

function notRunning(error: unknown): boolean {
	const code = (error as NodeJS.ErrnoException).code;
	return code === "ENOENT" || code === "ECONNREFUSED";
}

async function start(where: Paths): Promise<void> {
	mkdirSync(where.home, { recursive: true, mode: 0o700 });
	const log = openSync(where.log, "a", 0o600);
	const script = process.argv[1] ?? "";
	const child = spawn(process.execPath, [script, "daemon"], { detached: true, stdio: ["ignore", log, log] });
	let exited = false;
	child.on("exit", () => {
		exited = true;
	});
	child.unref();
	const deadline = Date.now() + START_TIMEOUT_MS;
	while (Date.now() < deadline && !exited) {
		if (await reachable(where.socket)) {
			return;
		}
		await new Promise((resolve) => setTimeout(resolve, 250));
	}
	if (await reachable(where.socket)) {
		return;
	}
	throw new Error(`the browser did not start; see ${where.log}`);
}
