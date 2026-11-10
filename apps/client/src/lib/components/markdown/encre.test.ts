import { readFileSync } from 'node:fs';
import { beforeAll, describe, expect, it } from 'vitest';
import { loadEncre, parseMarkdown, serializeMarkdown, validateDocument } from '../../../../node_modules/@facile/encre/src/lib/wasm';
import type { EncreNode } from '../../../../node_modules/@facile/encre/src/lib/schema';
import { toEncreMarkdown, type AlertTitles } from './encreCompat';
import { stripFrontmatter } from './MarkdownMuseUtils';

const titles: AlertTitles = { note: 'Note', tip: 'Tip', important: 'Important', warning: 'Warning', caution: 'Caution', pros: 'Advantages', cons: 'Drawbacks' };

const legacy = [
	'# Weekly',
	'',
	'```chart:bar',
	'title: Sales',
	'Jan: 10',
	'Feb: 25 k',
	'```',
	'',
	'```chart:metrics',
	'label: Repos',
	'val: 42',
	'---',
	'label: Active',
	'val: 16',
	'sub: In production',
	'```',
	'',
	'> [!WARNING]',
	'> Mind the **gap**',
	'',
	'```mermaid',
	'flowchart LR',
	'  A-->B',
	'```',
	'',
	'A [badge:success kept] line.',
	''
].join('\n');

function render(markdown: string): { mapped: string; doc: EncreNode; reason: string | null } {
	const mapped = toEncreMarkdown(stripFrontmatter(markdown), titles);
	const doc = parseMarkdown(mapped);
	return { mapped, doc, reason: validateDocument(doc) };
}

const blocks = (doc: EncreNode) => doc.content ?? [];
const attrs = (node: EncreNode) => node.attrs ?? {};

describe('encre rendering', () => {
	beforeAll(async () => {
		await loadEncre(readFileSync(new URL('../../../../node_modules/@facile/encre/src/lib/encre-wasm/encre_wasm_bg.wasm', import.meta.url)));
	});

	it('maps legacy fences and alerts onto encre blocks and keeps the rest', () => {
		const { doc, reason } = render(legacy);
		expect(reason).toBeNull();
		expect(blocks(doc).map((node) => node.type)).toEqual(['heading', 'chart', 'stat', 'stat', 'callout', 'codeBlock', 'paragraph']);
		expect(attrs(blocks(doc)[1])).toMatchObject({ kind: 'bar', title: 'Sales', labels: ['Jan', 'Feb'], series: [{ name: 'Sales', data: [10, 25] }] });
		expect(attrs(blocks(doc)[3])).toMatchObject({ label: 'Active', value: '16', delta: 'In production' });
		expect(attrs(blocks(doc)[4])).toMatchObject({ tone: 'warning', title: 'Warning' });
		expect(attrs(blocks(doc)[5]).language).toBe('mermaid');
		expect(blocks(doc)[6].content?.map((node) => node.text).join('')).toBe('A [badge:success kept] line.');
	});

	it('leaves an encre fence alone', () => {
		const source = '```chart\n{"kind": "line", "labels": ["a"], "series": [{"name": "s", "data": [1]}]}\n```\n';
		const { mapped, doc } = render(source);
		expect(mapped).toBe(source);
		expect(attrs(blocks(doc)[0]).kind).toBe('line');
	});

	it('saves through the canonical writer and validates again', () => {
		const { doc } = render(legacy);
		const saved = serializeMarkdown(doc);
		const again = parseMarkdown(saved);
		expect(again).toEqual(doc);
		expect(validateDocument(again)).toBeNull();
		expect(serializeMarkdown(again)).toBe(saved);
	});

	it('still opens an invalid artifact and names the reason', () => {
		const { doc, reason } = render('# Old\n\na\u0001b\n');
		expect(blocks(doc)).toHaveLength(2);
		expect(reason).toContain('forbidden');
	});
});
