import { createSseParser } from '$lib/sse';
import { parseLiveFrame, type LiveEvent } from './events';

export type LiveListener = {
	onEvent?: (event: LiveEvent) => void;
	onOpen?: () => void;
};

export type LiveDeps = {
	connect: (signal: AbortSignal) => Promise<Response>;
	wait: (ms: number) => Promise<void>;
};

const FIRST_BACKOFF_MS = 1000;
const MAX_BACKOFF_MS = 30000;

/**
 * One connection to the user's live stream, shared by every part of the app. It
 * reconnects with backoff and tells listeners each time it opens, so they can
 * refetch what they may have missed.
 */
export class LiveStream {
	private listeners = new Set<LiveListener>();
	private controller: AbortController | null = null;
	private deps: LiveDeps;

	constructor(deps: LiveDeps) {
		this.deps = deps;
	}

	get running(): boolean {
		return this.controller !== null;
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

	private async loop(signal: AbortSignal): Promise<void> {
		let backoff = FIRST_BACKOFF_MS;
		while (!signal.aborted) {
			const opened = await this.read(signal).catch(() => false);
			if (signal.aborted) return;
			backoff = opened ? FIRST_BACKOFF_MS : Math.min(backoff * 2, MAX_BACKOFF_MS);
			await this.deps.wait(backoff);
		}
	}

	private async read(signal: AbortSignal): Promise<boolean> {
		const response = await this.deps.connect(signal);
		if (!response.ok || !response.body) return false;
		for (const l of this.listeners) l.onOpen?.();
		const parse = createSseParser((frame) => {
			const event = parseLiveFrame(frame);
			if (event) for (const l of this.listeners) l.onEvent?.(event);
		});
		const decoder = new TextDecoder();
		const reader = response.body.getReader();
		for (;;) {
			const { done, value } = await reader.read();
			if (done) return true;
			parse(decoder.decode(value, { stream: true }));
		}
	}
}
