import assert from "node:assert/strict";
import { test } from "node:test";

import { Idle, REFUSED_SOCKET_CODE, refuseSocket } from "../src/guards.ts";

const wait = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

test("a page's WebSocket is closed with a policy code", async () => {
	const closes: Array<{ code?: number; reason?: string } | undefined> = [];
	await refuseSocket({ close: async (options) => void closes.push(options) });
	assert.equal(closes.length, 1);
	assert.equal(closes[0]?.code, REFUSED_SOCKET_CODE);
});

test("idle fires once nothing touches it, and a touch postpones it", async () => {
	let fired = 0;
	const idle = new Idle(40, () => fired++);
	await wait(25);
	idle.touch();
	await wait(25);
	assert.equal(fired, 0);
	await wait(40);
	assert.equal(fired, 1);
	idle.stop();
});

test("stop cancels the countdown", async () => {
	let fired = 0;
	const idle = new Idle(20, () => fired++);
	idle.stop();
	await wait(40);
	assert.equal(fired, 0);
});
