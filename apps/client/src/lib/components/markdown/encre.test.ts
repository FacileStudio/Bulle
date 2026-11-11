import { readFileSync } from 'node:fs';
import { beforeAll, describe, expect, it } from 'vitest';
import { loadEncre, parseMarkdown, serializeMarkdown, validateDocument } from '../../../../node_modules/@facile/encre/src/lib/wasm';
import type { EncreNode } from '../../../../node_modules/@facile/encre/src/lib/schema';
import { toEncreMarkdown, type AlertTitles } from './encreCompat';
import { openArtifact, refusalOf } from './encreOpen';
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

const schema2 = [
	'# Launch',
	'',
	'```mermaid',
	'flowchart LR',
	'  A-->B',
	'```',
	'',
	'Ship [badge:success Shipped] then [status:pending] with ==x== marked.',
	'',
	'```compare',
	'### Pros',
	'- fast',
	'### Cons',
	'- costly',
	'```',
	'',
	'> [!TIP]',
	'> Read it',
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
		expect(blocks(doc).map((node) => node.type)).toEqual(['heading', 'chart', 'stat', 'stat', 'callout', 'diagram', 'paragraph']);
		expect(attrs(blocks(doc)[1])).toMatchObject({ kind: 'bar', title: 'Sales', labels: ['Jan', 'Feb'], series: [{ name: 'Sales', data: [10, 25] }] });
		expect(attrs(blocks(doc)[3])).toMatchObject({ label: 'Active', value: '16', delta: 'In production' });
		expect(attrs(blocks(doc)[4])).toMatchObject({ tone: 'warning', title: 'Warning' });
		expect(attrs(blocks(doc)[5]).language).toBe('mermaid');
		expect(blocks(doc)[6].content?.map((node) => node.type)).toEqual(['text', 'badge', 'text']);
	});

	it('reads every schema 2 construct natively and localizes the alert title', () => {
		const { doc, reason } = render(schema2);
		expect(reason).toBeNull();
		expect(blocks(doc).map((node) => node.type)).toEqual(['heading', 'diagram', 'paragraph', 'columns', 'callout']);
		expect(attrs(blocks(doc)[1])).toEqual({ language: 'mermaid' });
		expect(blocks(doc)[1].content?.[0].text).toBe('flowchart LR\n  A-->B');
		const inline = blocks(doc)[2].content ?? [];
		expect(inline.filter((node) => node.type === 'badge').map(attrs)).toEqual([
			{ dot: false, tone: 'success' },
			{ dot: true, tone: 'warning' }
		]);
		expect(inline.find((node) => node.marks?.some((mark) => mark.type === 'highlight'))?.text).toBe('x');
		expect(blocks(doc)[3].content?.map(attrs)).toEqual([
			{ title: 'Pros', tone: 'success' },
			{ title: 'Cons', tone: 'danger' }
		]);
		expect(attrs(blocks(doc)[4])).toMatchObject({ tone: 'success', title: 'Tip' });
	});

	it('leaves an encre fence alone', () => {
		const source = '```chart\n{"kind": "line", "labels": ["a"], "series": [{"name": "s", "data": [1]}]}\n```\n';
		const { mapped, doc } = render(source);
		expect(mapped).toBe(source);
		expect(attrs(blocks(doc)[0]).kind).toBe('line');
	});

	it('saves through the canonical writer and validates again', () => {
		for (const source of [legacy, schema2]) {
			const { doc } = render(source);
			const saved = serializeMarkdown(doc);
			const again = parseMarkdown(saved);
			expect(again).toEqual(doc);
			expect(validateDocument(again)).toBeNull();
			expect(serializeMarkdown(again)).toBe(saved);
		}
	});

	it('still opens an invalid artifact and names the reason', () => {
		const { doc, reason } = render('# Old\n\na\u0001b\n');
		expect(blocks(doc)).toHaveLength(2);
		expect(reason).toContain('forbidden');
	});

	it('declines a document encre cannot read instead of throwing, with its place', () => {
		const nested = (depth: number) => {
			const lines = ['# Deep', ''];
			for (let i = 0; i < depth; i++) lines.push('`'.repeat(depth + 3 - i) + `callout tone="info" title="L${i + 1}"`);
			lines.push('deep');
			for (let i = depth - 1; i >= 0; i--) lines.push('`'.repeat(depth + 3 - i));
			return lines.join('\n') + '\n';
		};
		expect(() => parseMarkdown(nested(9))).toThrow();
		const refused = openArtifact(nested(9));
		expect(refused.doc).toBeNull();
		expect(refused.reason).toBe('line 11, column 1: markdown could not be read: custom blocks nest more than 8 levels deep');
		const fine = openArtifact(nested(8));
		expect(fine.doc?.content?.[1]?.type).toBe('callout');
		expect(fine.reason).toBeNull();
		expect(openArtifact('# Old\n\na\u0001b\n')).toMatchObject({ reason: expect.stringContaining('forbidden') });
	});

	it('places a refusal before saving, as the API would', () => {
		const nested = (depth: number) => {
			const lines = ['# Deep', ''];
			for (let i = 0; i < depth; i++) lines.push('`'.repeat(depth + 3 - i) + `callout tone="info" title="L${i + 1}"`);
			lines.push('deep');
			for (let i = depth - 1; i >= 0; i--) lines.push('`'.repeat(depth + 3 - i));
			return lines.join('\n') + '\n';
		};
		expect(refusalOf(nested(9))).toBe('line 11, column 1: custom blocks nest more than 8 levels deep');
		expect(refusalOf('# Hi\n\na\u0001b\n')).toBe('line 3, column 1: text is empty or holds a forbidden character');
		expect(refusalOf(nested(8))).toBeNull();
		expect(refusalOf(legacy)).toBeNull();
	});
});
