import type { ToolCall } from './reducer';

export type ToolKind = 'commands' | 'files' | 'web' | 'other';

export const KIND_LABEL: Record<ToolKind, string> = {
	commands: 'Commands',
	files: 'Files',
	web: 'Web',
	other: 'Other tools'
};

export function toolKind(name: string): ToolKind {
	const n = name.toLowerCase();
	if (/(bash|shell|exec|command|terminal)/.test(n)) return 'commands';
	if (/(read|write|edit|file|glob|grep|ls|list|patch|dir)/.test(n)) return 'files';
	if (/(web|fetch|http|url|search|browse|curl)/.test(n)) return 'web';
	return 'other';
}

export type ToolGroup = { kind: ToolKind; label: string; calls: ToolCall[] };

export function groupTools(tools: ToolCall[]): ToolGroup[] {
	const groups = new Map<ToolKind, ToolGroup>();
	for (const call of tools) {
		const kind = toolKind(call.name);
		const group = groups.get(kind) ?? { kind, label: KIND_LABEL[kind], calls: [] };
		group.calls.push(call);
		groups.set(kind, group);
	}
	return [...groups.values()];
}

const SUMMARY_KEYS = ['command', 'cmd', 'path', 'file_path', 'url', 'query', 'pattern'];

export function summarizeArgs(args: unknown, max = 90): string {
	let text = '';
	if (typeof args === 'string') text = args;
	else if (args && typeof args === 'object') {
		const record = args as Record<string, unknown>;
		const key = SUMMARY_KEYS.find((k) => typeof record[k] === 'string');
		text = key ? String(record[key]) : JSON.stringify(args);
	} else if (args !== undefined && args !== null) text = String(args);
	text = text.replace(/\s+/g, ' ').trim();
	return text.length > max ? `${text.slice(0, max - 1)}…` : text;
}

export function prettyArgs(args: unknown): string {
	if (typeof args === 'string') return args;
	if (args === undefined || args === null) return '';
	return JSON.stringify(args, null, 2);
}

export function runningCount(group: ToolGroup): number {
	return group.calls.filter((c) => c.status === 'running').length;
}

export function failedCount(group: ToolGroup): number {
	return group.calls.filter((c) => c.status === 'failed').length;
}
