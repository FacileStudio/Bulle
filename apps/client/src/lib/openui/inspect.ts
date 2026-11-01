import { createParser } from '@openuidev/lang-core';
import type { Library } from '@openuidev/svelte-lang';

export type Inspection =
	| { status: 'ready' }
	| { status: 'pending' }
	| { status: 'fallback'; reason: string };

const STREAMING_FATAL = new Set(['unknown-component', 'inline-reserved']);

/**
 * Decides whether an OpenUI source can be rendered. While streaming only
 * unknown components are fatal, because missing props are still on the way.
 * Once complete, any parse error sends the caller to the raw-source fallback.
 */
export function createInspector(library: Library) {
	const parser = createParser(library.toJSONSchema());

	return function inspect(source: string, streaming: boolean): Inspection {
		if (source.trim() === '') return streaming ? { status: 'pending' } : { status: 'fallback', reason: 'empty' };

		let result;
		try {
			result = parser.parse(source);
		} catch (error) {
			return { status: 'fallback', reason: error instanceof Error ? error.message : 'parse failed' };
		}

		const fatal = result.meta.errors.find((e) => !streaming || STREAMING_FATAL.has(e.code));
		if (fatal) return { status: 'fallback', reason: fatal.message };
		if (!result.root) return streaming ? { status: 'pending' } : { status: 'fallback', reason: 'no root' };
		return { status: 'ready' };
	};
}
