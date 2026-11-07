import { describe, expect, it } from 'vitest';
import { applyRichFormats, richBlock, richInline } from './MarkdownMuseRich';
import {
	fenceKind,
	parseAlert,
	parseBarChart,
	parseCompare,
	parseListItem,
	parseMetrics,
	stripFrontmatter
} from './MarkdownMuseUtils';

describe('markdown utils', () => {
	it('strips frontmatter', () => {
		expect(stripFrontmatter('---\ntitle: x\n---\n# Hi')).toBe('# Hi');
		expect(stripFrontmatter('# Hi')).toBe('# Hi');
	});

	it('parses alerts', () => {
		const alert = parseAlert('[!WARNING]\nMind the gap');
		expect(alert).toMatchObject({ isAlert: true, tone: 'warning', title: 'Warning', body: 'Mind the gap' });
		expect(parseAlert('[!PROS] fast').tone).toBe('pros');
		expect(parseAlert('just a quote').isAlert).toBe(false);
	});

	it('parses bar charts', () => {
		const chart = parseBarChart('title: Sales\nJan: 10\nFeb: 25 k');
		expect(chart.title).toBe('Sales');
		expect(chart.items.map((i) => i.value)).toEqual([10, 25]);
		expect(chart.max).toBe(25);
	});

	it('parses metrics', () => {
		expect(parseMetrics('label: Users\nvalue: 120\n---\nlabel: Churn\nval: 2%\nsub: monthly')).toEqual([
			{ label: 'Users', val: '120', sub: '' },
			{ label: 'Churn', val: '2%', sub: 'monthly' }
		]);
	});

	it('parses compare sections', () => {
		const sections = parseCompare('### Advantages\n- fast\n### Drawbacks\n- costly\n- loud');
		expect(sections.map((s) => s.tone)).toEqual(['pro', 'con']);
		expect(sections[1].items).toHaveLength(2);
		expect(parseCompare('Plain\n---\nOther').map((s) => s.title)).toEqual(['Plain', 'Other']);
	});

	it('reads list markers', () => {
		expect(parseListItem('[+] quick')).toEqual({ kind: 'pro', clean: 'quick' });
		expect(parseListItem('[!] careful').kind).toBe('warn');
		expect(parseListItem('plain').kind).toBe('normal');
	});

	it('maps fence tags', () => {
		expect(fenceKind('chart:bar')).toBe('bars');
		expect(fenceKind('chart:metrics')).toBe('metrics');
		expect(fenceKind('vs')).toBe('compare');
		expect(fenceKind('ts')).toBe('code');
	});
});

describe('rich rendering', () => {
	it('escapes raw html and unsafe links', () => {
		expect(richBlock('<img src=x onerror=alert(1)>')).not.toContain('<img');
		expect(richInline('[x](javascript:alert(1))')).not.toContain('href');
	});

	it('adds badges and highlights outside code', () => {
		const out = richInline('[badge:success Done] ==key== `==raw==`');
		expect(out).toContain('bg-fc-success/15');
		expect(out).toContain('<mark');
		expect(out).toContain('<code>==raw==</code>');
	});

	it('never decorates inside a tag', () => {
		expect(applyRichFormats('<a href="x==y==">z</a>')).toBe('<a href="x==y==">z</a>');
	});
});
