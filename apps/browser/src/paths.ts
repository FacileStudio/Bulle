import { homedir } from "node:os";
import { join } from "node:path";

/** Where the daemon keeps its socket, profile and log. */
export interface Paths {
	home: string;
	socket: string;
	profile: string;
	log: string;
}

/** The daemon's files under ~/.bulle, or under BULLE_BROWSER_HOME when set. */
export function paths(): Paths {
	const home = process.env.BULLE_BROWSER_HOME ?? join(homedir(), ".bulle");
	return {
		home,
		socket: join(home, "browser.sock"),
		profile: join(home, "browser"),
		log: join(home, "browser.log"),
	};
}
