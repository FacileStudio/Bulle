import { connect, type Socket } from "node:net";

import type { Args } from "./commands.ts";
import type { Reply } from "./session.ts";

/** One request line from the CLI to the daemon. */
export interface Request {
	cmd: string;
	args: Args;
}

/** One answer line from the daemon to the CLI. */
export type Response = ({ ok: true } & Reply) | { ok: false; error: string };

/** Reads one newline-terminated JSON value from a socket. */
export function readLine<T>(socket: Socket): Promise<T> {
	return new Promise((resolve, reject) => {
		let buffer = "";
		socket.setEncoding("utf8");
		socket.on("data", (chunk: string) => {
			buffer += chunk;
			const end = buffer.indexOf("\n");
			if (end >= 0) {
				resolve(JSON.parse(buffer.slice(0, end)) as T);
			}
		});
		socket.on("error", reject);
		socket.on("end", () => reject(new Error("the daemon closed the connection without an answer")));
	});
}

/** Sends one request to the daemon at socketPath and waits for its answer. */
export function send(socketPath: string, request: Request): Promise<Response> {
	return new Promise((resolve, reject) => {
		const socket = connect(socketPath);
		socket.on("error", reject);
		socket.on("connect", () => {
			socket.write(`${JSON.stringify(request)}\n`);
			readLine<Response>(socket).then(resolve, reject).finally(() => socket.end());
		});
	});
}

/** Whether a daemon answers on socketPath. */
export async function reachable(socketPath: string): Promise<boolean> {
	return send(socketPath, { cmd: "status", args: {} }).then(
		() => true,
		() => false,
	);
}
