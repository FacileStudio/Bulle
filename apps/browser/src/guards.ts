/** How long the daemon waits for a command before it exits. */
export const IDLE_MS = 15 * 60_000;

/** The close code a page sees for a refused WebSocket: policy violation. */
export const REFUSED_SOCKET_CODE = 1008;

/** The slice of a Playwright WebSocketRoute that refusing needs. */
export interface ClosableSocket {
	close(options?: { code?: number; reason?: string }): Promise<void>;
}

/**
 * Refuses a WebSocket a page opened. A socket is a second channel out of the
 * browser that no held request would ever see, so none is let through.
 */
export async function refuseSocket(socket: ClosableSocket): Promise<void> {
	await socket.close({ code: REFUSED_SOCKET_CODE, reason: "bulle-browser blocks WebSockets" }).catch(() => undefined);
}

/**
 * Calls onIdle once no command has touched it for ms. The daemon exits then,
 * keeping the profile on disk, and the next command starts a fresh one.
 */
export class Idle {
	#timer: ReturnType<typeof setTimeout> | undefined;
	#ms: number;
	#onIdle: () => void;

	constructor(ms: number, onIdle: () => void) {
		this.#ms = ms;
		this.#onIdle = onIdle;
		this.touch();
	}

	/** Restarts the countdown. */
	touch(): void {
		clearTimeout(this.#timer);
		this.#timer = setTimeout(this.#onIdle, this.#ms);
	}

	/** Cancels the countdown for good. */
	stop(): void {
		clearTimeout(this.#timer);
	}
}
