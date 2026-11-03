import { playwrightVersion } from "./playwright.ts";
import type { Reply, Session } from "./session.ts";

/** A command's JSON arguments, as the CLI received them. */
export type Args = Record<string, unknown>;

type Handler = (session: Session, args: Args) => Promise<Reply>;

/**
 * Commands that drive the page. They run one at a time, because two actions
 * racing on one tab leave a snapshot that belongs to neither.
 */
export const actions: Record<string, Handler> = {
	navigate: (session, args) =>
		session.act((page) => page.goto(webURL(field(args, "url")), { waitUntil: "domcontentloaded" })),
	snapshot: (session) => session.snapshot(),
	find: (session, args) => session.find(field(args, "text")),
	click: (session, args) => session.act((page) => page.getByRef(field(args, "ref")).click({ noWaitAfter: true })),
	type: (session, args) =>
		session.act(async (page) => {
			const target = page.getByRef(field(args, "ref"));
			await target.fill(field(args, "text"));
			if (args.submit === true) {
				await target.press("Enter", { noWaitAfter: true });
			}
		}),
	screenshot: (session, args) => session.screenshot(args.full_page === true),
	back: (session) => session.act((page) => page.goBack({ waitUntil: "domcontentloaded" })),
	confirm: async (session, args) => {
		const record = await session.holds.confirm(field(args, "held_id"));
		const reply = await session.act(async () => undefined);
		return { ...reply, text: `Sent ${record.method} ${record.url}.\n${reply.text}` };
	},
};

/** Commands that answer at once, even while an action is still running. */
export const controls: Record<string, Handler> = {
	cancel: async (session, args) => {
		const record = await session.holds.cancel(field(args, "held_id"));
		return { text: `Dropped ${record.method} ${record.url}; it was never sent.`, held: session.holds.list() };
	},
	abort: async (session) => {
		const dropped = await session.holds.abortAll();
		await session.stopLoading();
		return { text: `Stopped. ${dropped} held requests dropped.`, held: [] };
	},
	status: async (session) => {
		const held = session.holds.list();
		const page = await session.page();
		return { text: `Running on ${page.url()} with ${held.length} held requests.`, held, url: page.url() };
	},
};

/** The plain-text value of a required argument. */
export function field(args: Args, name: string): string {
	const value = args[name];
	if (typeof value !== "string" || value.trim() === "") {
		throw new Error(`${name} is required`);
	}
	return value;
}

/** A URL the browser may open: http or https only. */
export function webURL(raw: string): string {
	const url = new URL(raw);
	if (url.protocol !== "http:" && url.protocol !== "https:") {
		throw new Error(`only http and https URLs can be opened, not ${url.protocol}`);
	}
	return url.toString();
}

/** The version line Bulle's preflight and humans read. */
export function versionLine(version: string): string {
	return `bulle-browser ${version} playwright ${playwrightVersion()}`;
}
