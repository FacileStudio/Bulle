import {
	EncreDocumentError,
	EncreMarkdownError,
	parseMarkdown,
	validateDocument,
	validateMarkdown,
	type EncreNode,
	type MarkdownProblem
} from '@facile/encre';

/** An artifact as the page or the editor opens it: the document, or the reason encre cannot read it. */
export type OpenedArtifact = { doc: EncreNode; reason: string | null } | { doc: null; reason: string };

/** A refusal as the API words it: `line L, column C: <message>` when encre knows the place. */
export function placed(message: string, line?: number, column?: number): string {
	return line ? `line ${line}, column ${column ?? 1}: ${message}` : message;
}

/**
 * Parses and validates an artifact body once encre is loaded. encre refuses markdown that nests
 * more than 64 levels or opens more than 8 custom blocks on one path by throwing an
 * `EncreMarkdownError` with its line and column: that becomes the placed reason, so the page shows
 * the source and the editor declines to open it instead of crashing.
 */
export function openArtifact(body: string): OpenedArtifact {
	try {
		const doc = parseMarkdown(body);
		return { doc, reason: validateDocument(doc) };
	} catch (error) {
		if (error instanceof EncreMarkdownError) return { doc: null, reason: placed(error.message, error.line, error.column) };
		return { doc: null, reason: error instanceof Error ? error.message : String(error) };
	}
}

/**
 * Checks markdown the way the API will before it is sent, and answers the first refusal placed
 * like the API's, or null when encre accepts it. The server check stays the authority.
 */
export function refusalOf(markdown: string): string | null {
	const verdict = validateMarkdown(markdown);
	if (verdict.valid) return null;
	const first: MarkdownProblem | undefined = verdict.errors[0];
	return first ? placed(first.message, first.line, first.column) : 'invalid document';
}

/** What the editor should do with an error encre handed it. */
export type EditorRefusal =
	| { kind: 'document'; path: string; blockId?: string }
	| { kind: 'paste'; reason: string }
	| { kind: 'other'; reason: string };

/**
 * Sorts an editor error: an `EncreDocumentError` names the node Save cannot write (the app
 * reveals it), an `EncreMarkdownError` is a paste encre could not read and inserted as plain
 * text (the app says where, placed like the API), anything else is shown as it is.
 */
export function editorRefusal(error: unknown): EditorRefusal {
	if (error instanceof EncreDocumentError) return { kind: 'document', path: error.path, blockId: error.blockId };
	if (error instanceof EncreMarkdownError) return { kind: 'paste', reason: placed(error.message, error.line, error.column) };
	return { kind: 'other', reason: error instanceof Error ? error.message : String(error) };
}
