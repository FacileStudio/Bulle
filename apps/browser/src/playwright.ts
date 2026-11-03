import { existsSync } from "node:fs";
import { createRequire } from "node:module";
import { join } from "node:path";

import type { BrowserType } from "playwright";

/** Where the bake installs the Playwright library. */
export const PLAYWRIGHT_DIR = process.env.BULLE_PLAYWRIGHT_DIR ?? "/opt/playwright";

const BAKED_BROWSERS = "/opt/ms-playwright";

/**
 * Loads Playwright from the baked install instead of from next to this
 * bundle, which ships alone. createRequire keeps the lookup explicit, since
 * ESM imports ignore NODE_PATH.
 */
export function loadChromium(): BrowserType {
	if (!process.env.PLAYWRIGHT_BROWSERS_PATH && existsSync(BAKED_BROWSERS)) {
		process.env.PLAYWRIGHT_BROWSERS_PATH = BAKED_BROWSERS;
	}
	const require = createRequire(join(PLAYWRIGHT_DIR, "package.json"));
	return require("playwright").chromium as BrowserType;
}

/** The installed Playwright version, or "missing". */
export function playwrightVersion(): string {
	try {
		const require = createRequire(join(PLAYWRIGHT_DIR, "package.json"));
		return (require("playwright/package.json") as { version: string }).version;
	} catch {
		return "missing";
	}
}
