export type Segment =
	| { type: 'markdown'; text: string }
	| { type: 'openui'; source: string; complete: boolean };

const OPEN = /^ {0,3}(`{3,}|~{3,})[ \t]*([^\s`]*)[^`]*$/;
const OPENUI_INFO = /^openui(-lang)?$/i;
const PARTIAL_OPEN = /^ {0,3}`{3}(o(p(e(n(u)?)?)?)?)?$/i;

function closes(line: string, marker: string) {
	const trimmed = line.trim();
	return trimmed.length >= marker.length && trimmed === marker[0]!.repeat(trimmed.length);
}

/**
 * Splits assistant text into markdown and OpenUI segments on ```openui fences.
 * An unterminated openui fence yields a segment with complete=false, which is
 * how a block that is still streaming looks. With streaming=true a trailing
 * partial opening fence such as "```ope" is held back instead of flashing as text.
 */
export function splitFences(text: string, options: { streaming?: boolean } = {}): Segment[] {
	const lines = text.split('\n');
	const segments: Segment[] = [];
	let markdown: string[] = [];
	let block: string[] | null = null;
	let marker = '';
	let foreign = '';

	const flushMarkdown = () => {
		const joined = markdown.join('\n');
		if (joined.trim() !== '') segments.push({ type: 'markdown', text: joined });
		markdown = [];
	};

	for (let i = 0; i < lines.length; i++) {
		const line = lines[i]!;
		const last = i === lines.length - 1;

		if (block) {
			if (closes(line, marker)) {
				segments.push({ type: 'openui', source: block.join('\n'), complete: true });
				block = null;
			} else {
				block.push(line);
			}
			continue;
		}

		if (foreign) {
			markdown.push(line);
			if (closes(line, foreign)) foreign = '';
			continue;
		}

		if (last && options.streaming && PARTIAL_OPEN.test(line)) continue;

		const open = OPEN.exec(line);
		if (open && OPENUI_INFO.test(open[2]!)) {
			flushMarkdown();
			block = [];
			marker = open[1]!;
			continue;
		}
		if (open) foreign = open[1]!;
		markdown.push(line);
	}

	if (block) segments.push({ type: 'openui', source: block.join('\n'), complete: false });
	else flushMarkdown();
	return segments;
}
