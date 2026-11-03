import { chmod, mkdir, rm } from "node:fs/promises";
import { createServer, type Socket } from "node:net";

import { actions, controls, type Args } from "./commands.ts";
import { IDLE_MS, Idle } from "./guards.ts";
import { Holds } from "./holds.ts";
import type { Paths } from "./paths.ts";
import { Session, type Reply } from "./session.ts";
import { reachable, readLine, type Request, type Response } from "./wire.ts";

/** Runs one command and answers it. */
export type Runner = (cmd: string, args: Args) => Promise<Reply>;

/** Runs commands against session: actions one after another, controls at once. */
export function dispatcher(session: Session): Runner {
	let tail: Promise<unknown> = Promise.resolve();
	return (cmd, args) => {
		const control = controls[cmd];
		if (control) {
			return control(session, args);
		}
		const action = actions[cmd];
		if (!action) {
			return Promise.reject(new Error(`unknown command ${cmd}`));
		}
		const run = tail.then(() => action(session, args));
		tail = run.catch(() => undefined);
		return run;
	};
}

/** Serves the browser on the unix socket until SIGTERM, SIGINT or IDLE_MS without a command. */
export async function serve(where: Paths): Promise<void> {
	await mkdir(where.home, { recursive: true, mode: 0o700 });
	if (await reachable(where.socket)) {
		return;
	}
	await rm(where.socket, { force: true });
	const session = await Session.open(where.profile, new Holds());
	const shutdown = () => {
		server.close();
		void session.close().finally(() => process.exit(0));
	};
	const run = touching(dispatcher(session), new Idle(IDLE_MS, shutdown));
	const server = createServer((socket) => void answer(socket, run));
	server.listen(where.socket, () => void chmod(where.socket, 0o600));
	process.on("SIGTERM", shutdown);
	process.on("SIGINT", shutdown);
}

/** Wraps run so every command, and the end of every command, restarts the idle countdown. */
export function touching(run: Runner, idle: Idle): Runner {
	return (cmd, args) => {
		idle.touch();
		return run(cmd, args).finally(() => idle.touch());
	};
}

async function answer(socket: Socket, run: Runner): Promise<void> {
	let response: Response;
	try {
		const request = await readLine<Request>(socket);
		response = { ok: true, ...(await run(request.cmd, request.args ?? {})) };
	} catch (error) {
		response = { ok: false, error: error instanceof Error ? error.message : String(error) };
	}
	socket.end(`${JSON.stringify(response)}\n`);
}
