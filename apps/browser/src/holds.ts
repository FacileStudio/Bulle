/** How long a held request waits for the owner before it is dropped. */
export const HOLD_TIMEOUT_MS = 120_000;

const PREVIEW_CHARS = 2048;
const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);

/** The daemon's own record of a request it held, shown on the approval card. */
export interface HeldRecord {
	id: string;
	method: string;
	url: string;
	origin: string;
	body_preview: string;
	body_bytes: number;
	resource_type: string;
	held_at: string;
}

/** The slice of a Playwright Request that holding needs. */
export interface HoldableRequest {
	method(): string;
	url(): string;
	postData(): string | null;
	resourceType(): string;
}

/** The slice of a Playwright Route that holding needs. */
export interface HoldableRoute {
	request(): HoldableRequest;
	continue(): Promise<void>;
	abort(errorCode?: string): Promise<void>;
}

interface Entry {
	record: HeldRecord;
	route: HoldableRoute;
	timer: ReturnType<typeof setTimeout>;
}

/**
 * Holds every request that could change something on a server (anything but
 * GET, HEAD and OPTIONS) until the owner confirms or cancels it. A held route
 * stays open rather than aborted, so confirming sends the original request
 * unchanged. A hold nobody answers is aborted after the timeout.
 */
export class Holds {
	#entries = new Map<string, Entry>();
	#next = 1;
	#timeoutMs: number;
	#now: () => Date;

	constructor(timeoutMs = HOLD_TIMEOUT_MS, now: () => Date = () => new Date()) {
		this.#timeoutMs = timeoutMs;
		this.#now = now;
	}

	/** Routes one request: safe methods continue, the rest are held. */
	async handle(route: HoldableRoute): Promise<void> {
		if (SAFE_METHODS.has(route.request().method().toUpperCase())) {
			await settle(route.continue());
			return;
		}
		const record = this.#record(route.request());
		const timer = setTimeout(() => void this.cancel(record.id).catch(() => undefined), this.#timeoutMs);
		this.#entries.set(record.id, { record, route, timer });
	}

	/** Every request currently held, oldest first. */
	list(): HeldRecord[] {
		return [...this.#entries.values()].map((entry) => entry.record);
	}

	/** Sends a held request as it was fired. */
	async confirm(id: string): Promise<HeldRecord> {
		const entry = this.#take(id);
		await settle(entry.route.continue());
		return entry.record;
	}

	/** Drops a held request; the page sees a failed request. */
	async cancel(id: string): Promise<HeldRecord> {
		const entry = this.#take(id);
		await settle(entry.route.abort("aborted"));
		return entry.record;
	}

	/** Drops every held request and reports how many there were. */
	async abortAll(): Promise<number> {
		const ids = [...this.#entries.keys()];
		await Promise.all(ids.map((id) => this.cancel(id)));
		return ids.length;
	}

	#take(id: string): Entry {
		const entry = this.#entries.get(id);
		if (!entry) {
			throw new Error(`no held request ${id}: it was already sent, cancelled or timed out`);
		}
		clearTimeout(entry.timer);
		this.#entries.delete(id);
		return entry;
	}

	#record(request: HoldableRequest): HeldRecord {
		const body = request.postData() ?? "";
		return {
			id: `h${this.#next++}`,
			method: request.method().toUpperCase(),
			url: request.url(),
			origin: originOf(request.url()),
			body_preview: body.slice(0, PREVIEW_CHARS),
			body_bytes: Buffer.byteLength(body),
			resource_type: request.resourceType(),
			held_at: this.#now().toISOString(),
		};
	}
}

/** The origin of a URL, or the URL itself when it has none. */
export function originOf(url: string): string {
	try {
		return new URL(url).origin;
	} catch {
		return url;
	}
}

async function settle(action: Promise<void>): Promise<void> {
	await action.catch(() => undefined);
}
