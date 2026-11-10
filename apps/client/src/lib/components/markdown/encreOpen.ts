import { parseMarkdown, validateDocument, type EncreNode } from '@facile/encre';

/** An artifact as the page or the editor opens it: the document, or the reason encre cannot read it. */
export type OpenedArtifact = { doc: EncreNode; reason: string | null } | { doc: null; reason: string };

/**
 * Parses and validates an artifact body once encre is loaded. encre refuses markdown that nests
 * more than 64 levels or opens more than 8 custom blocks on one path by throwing: that becomes the
 * reason, with its line and column, so the page shows the source and the editor declines to open
 * it instead of crashing.
 */
export function openArtifact(body: string): OpenedArtifact {
	try {
		const doc = parseMarkdown(body);
		return { doc, reason: validateDocument(doc) };
	} catch (error) {
		return { doc: null, reason: error instanceof Error ? error.message : String(error) };
	}
}
