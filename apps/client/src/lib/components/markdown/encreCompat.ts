import { marked, type Token, type Tokens } from 'marked';
import { t } from '$lib/i18n/index.svelte';
import { parseBarChart, parseMetrics } from './MarkdownMuseUtils';

/** The titles the viewer gives the alert kinds, in the reader's locale. */
export type AlertTitles = Record<'note' | 'tip' | 'important' | 'warning' | 'caution' | 'pros' | 'cons', string>;

type Tone = 'neutral' | 'info' | 'success' | 'warning' | 'danger';

const ALERT = /^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION|PROS|CONS|ADVANTAGE|DRAWBACK)\]\s*/i;

const ALERT_KINDS: Record<string, { tone: Tone; title: keyof AlertTitles }> = {
	NOTE: { tone: 'info', title: 'note' },
	TIP: { tone: 'success', title: 'tip' },
	PROS: { tone: 'success', title: 'pros' },
	ADVANTAGE: { tone: 'success', title: 'pros' },
	IMPORTANT: { tone: 'info', title: 'important' },
	WARNING: { tone: 'warning', title: 'warning' },
	CAUTION: { tone: 'danger', title: 'caution' },
	CONS: { tone: 'danger', title: 'cons' },
	DRAWBACK: { tone: 'danger', title: 'cons' }
};

/** The alert titles in the active locale. */
export function alertTitles(): AlertTitles {
	return {
		note: t('markdown.alert.note'),
		tip: t('markdown.alert.tip'),
		important: t('markdown.alert.important'),
		warning: t('markdown.alert.warning'),
		caution: t('markdown.alert.caution'),
		pros: t('markdown.alert.pros'),
		cons: t('markdown.alert.cons')
	};
}

/**
 * Rewrites the rich fences Mycelium artifacts used before encre into encre's blocks, when the
 * page renders and never in Mycelium: `chart:bar` (or a `chart` whose body is not JSON) becomes
 * a `chart`, each `chart:metrics` card a `stat`, and a `> [!NOTE]`-style alert a `callout`.
 * Every other byte of the body is kept as it is.
 *
 * It stays with encre 0.8.0: encre reads GFM alerts natively but gives the callout an empty
 * title, and this rewrite is what puts the reader's "Note" or "Astuce" on it; `chart:bar` and
 * `chart:metrics` still have no encre reader. Nothing else belongs here: `mermaid`, `diff`,
 * `compare`, `[badge:..]`, `[status:..]` and `==mark==` reach encre untouched.
 */
export function toEncreMarkdown(body: string, titles: AlertTitles): string {
	const tokens = marked.lexer(body, { gfm: true }) as Token[];
	return tokens.map((token) => rewrite(token, titles) ?? token.raw).join('');
}

function rewrite(token: Token, titles: AlertTitles): string | null {
	if (token.type === 'code') return rewriteFence(token as Tokens.Code);
	if (token.type === 'blockquote') return rewriteAlert(token as Tokens.Blockquote, titles);
	return null;
}

function rewriteFence(code: Tokens.Code): string | null {
	const lang = (code.lang ?? '').trim().toLowerCase();
	if (lang.startsWith('chart:bar') || (lang === 'chart' && !code.text.trimStart().startsWith('{'))) {
		const chart = parseBarChart(code.text);
		if (chart.items.length === 0) return null;
		return fence(
			'chart',
			JSON.stringify(
				withValues({
					kind: 'bar',
					title: chart.title,
					labels: chart.items.map((item) => item.label),
					series: [{ name: chart.title, data: chart.items.map((item) => item.value) }]
				})
			)
		);
	}
	if (lang.startsWith('chart:metric')) {
		const metrics = parseMetrics(code.text);
		if (metrics.length === 0) return null;
		return metrics
			.map((metric) => fence('stat', JSON.stringify(withValues({ label: metric.label, value: metric.val, delta: metric.sub }))))
			.join('\n');
	}
	return null;
}

function rewriteAlert(quote: Tokens.Blockquote, titles: AlertTitles): string | null {
	const match = quote.text.match(ALERT);
	const kind = match && ALERT_KINDS[match[1].toUpperCase()];
	if (!match || !kind) return null;
	const body = quote.text.slice(match[0].length).trim();
	return fence('callout', body, ` tone="${attribute(kind.tone)}" title="${attribute(titles[kind.title])}"`);
}

function fence(name: string, body: string, attributes = ''): string {
	const ticks = '`'.repeat(Math.max(4, longestBacktickRun(body) + 1));
	return `${ticks}${name}${attributes}\n${body}\n${ticks}\n`;
}

function longestBacktickRun(text: string): number {
	let longest = 0;
	for (const run of text.matchAll(/`+/g)) longest = Math.max(longest, run[0].length);
	return longest;
}

function attribute(value: string): string {
	return value.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
}

function withValues<T extends Record<string, unknown>>(object: T): Partial<T> {
	return Object.fromEntries(Object.entries(object).filter(([, value]) => value !== '')) as Partial<T>;
}
