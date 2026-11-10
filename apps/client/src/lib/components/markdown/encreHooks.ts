import DOMPurify from 'dompurify';
import hljs from 'highlight.js/lib/common';
import type { Highlight, RenderDiagram } from '@facile/encre';
import { renderMermaid } from './mermaid';

/**
 * Draws a `mermaid` diagram with the app's mermaid, loaded on first use. `d2` and `dot` are
 * refused, so encre shows their source under a language label.
 */
export const renderDiagram: RenderDiagram = (language, source) =>
	language === 'mermaid' ? renderMermaid(source) : Promise.reject(new Error(`no renderer for ${language}`));

/**
 * Highlights a code block with highlight.js when it knows the language, else escapes the code.
 * encre inserts the result with `{@html}`, so it goes through DOMPurify first.
 */
export const highlight: Highlight = (code, language) =>
	DOMPurify.sanitize(
		hljs.getLanguage(language) ? hljs.highlight(code, { language, ignoreIllegals: true }).value : escapeHtml(code)
	);

function escapeHtml(text: string): string {
	return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
