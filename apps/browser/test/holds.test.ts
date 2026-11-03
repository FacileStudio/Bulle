import assert from "node:assert/strict";
import { test } from "node:test";

import { Holds, type HoldableRoute } from "../src/holds.ts";

function fakeRoute(method: string, url: string, body: string | null) {
	const calls: string[] = [];
	const route: HoldableRoute = {
		request: () => ({ method: () => method, url: () => url, postData: () => body, resourceType: () => "document" }),
		continue: async () => void calls.push("continue"),
		abort: async (code) => void calls.push(`abort:${code}`),
	};
	return { route, calls };
}

test("GET continues at once and is never held", async () => {
	const holds = new Holds();
	const get = fakeRoute("GET", "https://shop.test/", null);
	await holds.handle(get.route);
	assert.deepEqual(get.calls, ["continue"]);
	assert.equal(holds.list().length, 0);
});

test("POST is held with the daemon's record until confirmed", async () => {
	const holds = new Holds();
	const post = fakeRoute("POST", "https://shop.test/order?x=1", "qty=2");
	await holds.handle(post.route);
	assert.deepEqual(post.calls, []);
	const [record] = holds.list();
	assert.equal(record.method, "POST");
	assert.equal(record.url, "https://shop.test/order?x=1");
	assert.equal(record.origin, "https://shop.test");
	assert.equal(record.body_preview, "qty=2");
	await holds.confirm(record.id);
	assert.deepEqual(post.calls, ["continue"]);
	await assert.rejects(holds.confirm(record.id), /no held request/);
});

test("cancel and abortAll abort held requests", async () => {
	const holds = new Holds();
	const first = fakeRoute("DELETE", "https://a.test/1", null);
	const second = fakeRoute("PUT", "https://a.test/2", "{}");
	await holds.handle(first.route);
	await holds.handle(second.route);
	await holds.cancel("h1");
	assert.deepEqual(first.calls, ["abort:aborted"]);
	assert.equal(await holds.abortAll(), 1);
	assert.deepEqual(second.calls, ["abort:aborted"]);
	assert.equal(holds.list().length, 0);
});

test("an unanswered hold is aborted after the timeout", async () => {
	const holds = new Holds(20);
	const post = fakeRoute("POST", "https://a.test/", "x");
	await holds.handle(post.route);
	await new Promise((resolve) => setTimeout(resolve, 60));
	assert.deepEqual(post.calls, ["abort:aborted"]);
	assert.equal(holds.list().length, 0);
});
