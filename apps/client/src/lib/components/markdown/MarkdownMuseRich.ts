import { renderInline, renderMarkdown } from '$lib/chat/markdown';

const PILL = 'inline-flex items-center gap-1 rounded-fc-pill px-2 py-0.5 text-fc-xs font-semibold';

export const toneColors: Record<string, string> = {
	success: 'bg-fc-success/15 text-fc-success',
	danger: 'bg-fc-danger/15 text-fc-danger',
	warning: 'bg-fc-warning/15 text-fc-warning',
	info: 'bg-fc-info/15 text-fc-info',
	primary: 'bg-fc-accent/15 text-fc-accent',
	neutral: 'bg-fc-surface text-fc-fg-muted'
};

function status(tone: string) {
	return (_: string, label: string) =>
		`<span class="${PILL} ${toneColors[tone]}"><span class="size-1.5 rounded-full bg-current"></span>${label}</span>`;
}

const RULES: [RegExp, (match: string, ...groups: string[]) => string][] = [
	[
		/\[(?:badge|tag|chip):(success|danger|warning|info|primary|neutral)\s+([^\]]+)\]/gi,
		(_, tone, label) => `<span class="${PILL} ${toneColors[tone.toLowerCase()]}">${label}</span>`
	],
	[/\[\+\s+([^\]]+)\]/g, (_, label) => `<span class="${PILL} ${toneColors.success}">+ ${label}</span>`],
	[/\[-\s+([^\]]+)\]/g, (_, label) => `<span class="${PILL} ${toneColors.danger}">- ${label}</span>`],
	[/\[status:(active|done|ready|success)\]/gi, status('success')],
	[/\[status:(pending|wip|running|loading)\]/gi, status('warning')],
	[/\[status:(error|failed|stopped|dead)\]/gi, status('danger')],
	[/==([^=]+)==/g, (_, text) => `<mark class="rounded-fc-xs bg-fc-warning/20 px-1 text-fc-fg">${text}</mark>`]
];

function decorate(text: string): string {
	return RULES.reduce((acc, [pattern, render]) => acc.replace(pattern, render), text);
}

/**
 * Adds badges, status pills and highlights to already-sanitised HTML. Only text between tags
 * is touched, and never inside `code` or `pre`, so the markup the renderer escaped stays inert.
 */
export function applyRichFormats(html: string): string {
	let literal = 0;
	return html
		.split(/(<[^>]*>)/)
		.map((part) => {
			if (!part.startsWith('<')) return literal ? part : decorate(part);
			if (/^<(code|pre)[\s>]/i.test(part)) literal += 1;
			else if (/^<\/(code|pre)>/i.test(part)) literal = Math.max(0, literal - 1);
			return part;
		})
		.join('');
}

/** Renders one line of user markdown safely, then applies the rich inline formats. */
export function richInline(raw: string): string {
	return applyRichFormats(renderInline(raw));
}

/** Renders a markdown block safely, then applies the rich inline formats. */
export function richBlock(raw: string): string {
	return applyRichFormats(renderMarkdown(raw));
}
