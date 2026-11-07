import { Marked } from 'marked';

const SAFE_URL = /^(https?:|mailto:|#|\/)/i;

function escapeHtml(text: string): string {
	return text
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;');
}

const parser = new Marked({
	gfm: true,
	breaks: true,
	renderer: {
		html({ text }) {
			return escapeHtml(text);
		},
		link({ href, title, tokens }) {
			const label = this.parser.parseInline(tokens);
			if (!SAFE_URL.test(href)) return label;
			const hint = title ? ` title="${escapeHtml(title)}"` : '';
			return `<a href="${escapeHtml(href)}"${hint} target="_blank" rel="noopener noreferrer nofollow">${label}</a>`;
		},
		image({ href, text }) {
			if (!SAFE_URL.test(href)) return escapeHtml(text);
			return `<a href="${escapeHtml(href)}" target="_blank" rel="noopener noreferrer nofollow">${escapeHtml(text || href)}</a>`;
		}
	}
});

export function renderMarkdown(source: string): string {
	return parser.parse(source, { async: false });
}

/** Renders one line of markdown with the same escaping and link rules, without a wrapping paragraph. */
export function renderInline(source: string): string {
	return parser.parseInline(source, { async: false });
}
