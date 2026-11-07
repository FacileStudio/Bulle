import { t } from '$lib/i18n/index.svelte';

export type AlertTone = 'note' | 'tip' | 'important' | 'warning' | 'caution' | 'pros' | 'cons';

export type AlertData = {
	isAlert: boolean;
	tone: AlertTone;
	title: string;
	body: string;
};

export type ListKind = 'pro' | 'con' | 'warn' | 'info' | 'normal';

export type BarItem = { label: string; value: number; formatted: string };

export type Metric = { label: string; val: string; sub: string };

export type CompareSection = { title: string; tone: 'pro' | 'con' | 'neutral'; items: string[] };

export type DiffLine = { type: 'add' | 'del' | 'hunk' | 'ctx'; line: string };

/** Drops a leading YAML frontmatter block, so the artifact body starts at its first heading. */
export function stripFrontmatter(raw: string): string {
	if (!raw.startsWith('---\n') && !raw.startsWith('---\r\n')) return raw;
	const end = raw.indexOf('\n---', 4);
	if (end < 0) return raw;
	const cut = raw.indexOf('\n', end + 4);
	return cut >= 0 ? raw.slice(cut + 1) : '';
}

function splitPair(line: string): [string, string] | null {
	const colon = line.indexOf(':');
	if (colon < 0) return null;
	return [line.slice(0, colon).trim(), line.slice(colon + 1).trim()];
}

/** Reads a `chart:bar` fence: `title: X` plus one `label: value` line per bar. */
export function parseBarChart(raw: string): { title: string; items: BarItem[]; max: number } {
	let title = '';
	const items: BarItem[] = [];
	for (const line of raw.trim().split('\n')) {
		const pair = splitPair(line.trim());
		if (!pair) continue;
		const [key, value] = pair;
		if (key.toLowerCase() === 'title') {
			title = value;
			continue;
		}
		const num = value.match(/[\d.,]+/);
		items.push({ label: key, value: num ? parseFloat(num[0].replace(',', '.')) || 0 : 0, formatted: value });
	}
	return { title, items, max: Math.max(...items.map((i) => i.value), 1) };
}

const METRIC_KEYS: Record<string, keyof Metric> = {
	label: 'label',
	title: 'label',
	val: 'val',
	value: 'val',
	sub: 'sub',
	desc: 'sub',
	subtitle: 'sub'
};

/** Reads a `chart:metrics` fence: blocks split by `---`, each with label, value and sub lines. */
export function parseMetrics(raw: string): Metric[] {
	return raw
		.trim()
		.split(/---+/)
		.map((block) => {
			const metric: Metric = { label: '', val: '', sub: '' };
			for (const line of block.trim().split('\n')) {
				const pair = splitPair(line);
				const key = pair && METRIC_KEYS[pair[0].toLowerCase()];
				if (pair && key) metric[key] = pair[1];
			}
			return metric;
		})
		.filter((m) => m.val || m.label);
}

const ALERTS: Record<string, { tone: AlertTone; title: string }> = {
	NOTE: {
		tone: 'note',
		get title() {
			return t('markdown.alert.note');
		}
	},
	TIP: {
		tone: 'tip',
		get title() {
			return t('markdown.alert.tip');
		}
	},
	IMPORTANT: {
		tone: 'important',
		get title() {
			return t('markdown.alert.important');
		}
	},
	WARNING: {
		tone: 'warning',
		get title() {
			return t('markdown.alert.warning');
		}
	},
	CAUTION: {
		tone: 'caution',
		get title() {
			return t('markdown.alert.caution');
		}
	},
	PROS: {
		tone: 'pros',
		get title() {
			return t('markdown.alert.pros');
		}
	},
	ADVANTAGE: {
		tone: 'pros',
		get title() {
			return t('markdown.alert.pros');
		}
	},
	CONS: {
		tone: 'cons',
		get title() {
			return t('markdown.alert.cons');
		}
	},
	DRAWBACK: {
		tone: 'cons',
		get title() {
			return t('markdown.alert.cons');
		}
	}
};

/** Recognises a GitHub-style `> [!NOTE]` alert, plus the PROS and CONS variants. */
export function parseAlert(raw: string): AlertData {
	const trimmed = raw.trim();
	const match = trimmed.match(/^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION|PROS|CONS|ADVANTAGE|DRAWBACK)\]\s*(.*)$/im);
	if (!match || match.index !== 0) return { isAlert: false, tone: 'note', title: '', body: raw };
	const kind = ALERTS[match[1].toUpperCase()];
	const extra = match[2].trim();
	const rest = trimmed.slice(match[0].length).trim();
	const body = extra ? `${extra}\n${rest}`.trim() : rest;
	return { isAlert: true, tone: kind.tone, title: kind.title, body };
}

const LIST_MARKERS: { kind: Exclude<ListKind, 'normal'>; pattern: RegExp }[] = [
	{ kind: 'pro', pattern: /^(\[\+\]|\+\s+|✅|\[pros?\]|\[advantage\])\s*/i },
	{ kind: 'con', pattern: /^(\[-\]|-\s+|❌|\[cons?\]|\[drawback\])\s*/i },
	{ kind: 'warn', pattern: /^(\[!\]|⚠️|\[warn\])\s*/i },
	{ kind: 'info', pattern: /^(\[\?\]|ℹ️|\[info\])\s*/i }
];

/** Reads the marker at the start of a list item (`[+]`, `[-]`, `[!]`, `[?]`) and strips it. */
export function parseListItem(text: string): { kind: ListKind; clean: string } {
	const trimmed = text.trim();
	for (const { kind, pattern } of LIST_MARKERS) {
		if (pattern.test(trimmed)) return { kind, clean: trimmed.replace(pattern, '') };
	}
	return { kind: 'normal', clean: text };
}

/** Classifies each line of a `diff` fence as added, removed, hunk header or context. */
export function parseDiff(text: string): DiffLine[] {
	return text.split('\n').map((line) => {
		if (line.startsWith('+') && !line.startsWith('+++')) return { type: 'add', line };
		if (line.startsWith('-') && !line.startsWith('---')) return { type: 'del', line };
		if (line.startsWith('@@')) return { type: 'hunk', line };
		return { type: 'ctx', line };
	});
}

function section(title: string, items: string[]): CompareSection {
	const tone = /advantage|pro|positive|plus|benefit/i.test(title)
		? 'pro'
		: /drawback|con|negative|minus|risk|issue/i.test(title)
			? 'con'
			: 'neutral';
	return { title, tone, items };
}

function toSection(chunk: string): CompareSection {
	const lines = chunk.trim().split('\n');
	const items = lines
		.slice(1)
		.map((l) => l.trim())
		.filter(Boolean);
	return section(lines[0].replace(/^#+\s*/, '').trim() || t('markdown.section'), items);
}

/** Reads a `compare` fence: `### Title` sections, or blocks split by `---`, each with items. */
export function parseCompare(raw: string): CompareSection[] {
	const byHeading = raw.split(/^###\s+/m).filter((s) => s.trim());
	if (byHeading.length >= 2) return byHeading.map(toSection);
	const byRule = raw.split(/---+/).filter((s) => s.trim());
	if (byRule.length >= 2) return byRule.map(toSection);
	return [];
}

export type FenceKind = 'bars' | 'metrics' | 'compare' | 'mermaid' | 'code';

/** Maps a fence's language tag to the rich block that renders it. */
export function fenceKind(lang: string | undefined): FenceKind {
	const tag = (lang ?? '').toLowerCase().trim();
	if (tag === 'chart' || tag.startsWith('chart:bar')) return 'bars';
	if (tag.startsWith('chart:metric')) return 'metrics';
	if (tag === 'compare' || tag === 'pros-cons' || tag === 'vs') return 'compare';
	if (tag === 'mermaid') return 'mermaid';
	return 'code';
}
