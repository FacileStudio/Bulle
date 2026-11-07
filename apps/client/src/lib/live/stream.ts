import { createSseParser } from '$lib/sse';
import { parseLiveFrame, type LiveEvent } from './events';

export type LiveListener = {
	onEvent?: (event: LiveEvent) => void;
	onOpen?: () => void;
	onUnauthorized?: () => void;
};

export type LiveDeps = {
	connect: (signal: AbortSignal) => Promise<Response>;
	wait: (ms: number) => Promise<void>;
	pingMs?: number;
	now?: () => number;
};

type ReadResult = 'opened' | 'failed' | 'unauthorized';

const FIRST_BACKOFF_MS = 1000;
const MAX_BACKOFF_MS = 30000;
export const SERVER_PING_MS = 15000;
const WATCHDOG_FACTOR = 2.5;

/**
 * One connection to the user's live stream, shared by every part of the app. It
 * reconnects with backoff, drops a connection that stays silent for 2.5 server
 * pings, stops on a 401, and tells listeners each time it opens, so they can
 * refetch what they may have missed.
 */
export class LiveStream {
	private listeners = new Set<LiveListener>();
	private controller: AbortController | null = null;
	private deps: LiveDeps;
	private lastData = 0;

	constructor(deps: LiveDeps) {
		this.deps = deps;
	}

	get running(): boolean {
		return this.controller !== null;
	}

	private get pingMs(): number {
		return this.deps.pingMs ?? SERVER_PING_MS;
	}

	private now(): number {
		return (this.deps.now ?? Date.now)();
	}

	/** Registers a listener and returns the function that removes it. */
	listen(listener: LiveListener): () => void {
		this.listeners.add(listener);
		return () => this.listeners.delete(listener);
	}

	/** Opens the stream unless it is already open. */
	start(): void {
		if (this.controller) return;
		const controller = new AbortController();
		this.controller = controller;
		void this.loop(controller.signal);
	}

	/** Closes the stream and stops reconnecting. */
	stop(): void {
		this.controller?.abort();
		this.controller = null;
	}

	/**
	 * Reconnects now when the stream is running but has heard nothing for a ping
	 * interval, as after the tab or the app comes back from the background.
	 */
	revive(): void {
		if (!this.controller || this.now() - this.lastData <= this.pingMs) return;
		this.stop();
		this.start();
	}

	private async loop(signal: AbortSignal): Promise<void> {
		let backoff = FIRST_BACKOFF_MS;
		while (!signal.aborted) {
			const result = await this.read(signal);
			if (signal.aborted) return;
			if (result === 'unauthorized') {
				this.stop();
				for (const l of this.listeners) l.onUnauthorized?.();
				return;
			}
			backoff = result === 'opened' ? FIRST_BACKOFF_MS : Math.min(backoff * 2, MAX_BACKOFF_MS);
			await this.deps.wait(backoff);
		}
	}

	private async read(signal: AbortSignal): Promise<ReadResult> {
		const connection = new AbortController();
		const abort = () => connection.abort();
		signal.addEventListener('abort', abort);
		let watchdog: ReturnType<typeof setTimeout> | undefined;
		const feed = () => {
			this.lastData = this.now();
			clearTimeout(watchdog);
			watchdog = setTimeout(abort, this.pingMs * WATCHDOG_FACTOR);
		};
		let opened = false;
		try {
			const response = await this.deps.connect(connection.signal);
			if (response.status === 401) return 'unauthorized';
			if (!response.ok || !response.body) return 'failed';
			opened = true;
			feed();
			for (const l of this.listeners) l.onOpen?.();
			await this.drain(response.body, feed);
		} catch {
			return opened ? 'opened' : 'failed';
		} finally {
			clearTimeout(watchdog);
			signal.removeEventListener('abort', abort);
		}
		return 'opened';
	}

	private async drain(body: ReadableStream<Uint8Array>, feed: () => void): Promise<void> {
		const parse = createSseParser((frame) => {
			const event = parseLiveFrame(frame);
			if (event) for (const l of this.listeners) l.onEvent?.(event);
		});
		const decoder = new TextDecoder();
		const reader = body.getReader();
		for (;;) {
			const { done, value } = await reader.read();
			if (done) return;
			feed();
			parse(decoder.decode(value, { stream: true }));
		}
	}
}
