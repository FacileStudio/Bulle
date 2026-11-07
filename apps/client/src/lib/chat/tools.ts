import type { ToolCall } from "./reducer";
import { locale, t } from "$lib/i18n/index.svelte";
import type { AppKey } from "$lib/i18n/messages";
import { APPROVAL_TOOLS, PHRASE_TOOLS } from "$lib/i18n/messages/chat";

export type ToolKind = "apps" | "commands" | "files" | "web" | "other";

export function kindLabel(kind: ToolKind): string {
  return t(`tools.kind.${kind}`);
}

export function toolKind(name: string): ToolKind {
  const n = name.toLowerCase();
  if (/^(courrier|agenda|opus|sablier|glouton)_/.test(n)) return "apps";
  if (/(bash|shell|exec|command|terminal)/.test(n)) return "commands";
  if (/(read|write|edit|file|glob|grep|ls|list|patch|dir)/.test(n))
    return "files";
  if (/(web|fetch|http|url|search|browse|curl)/.test(n)) return "web";
  return "other";
}

export type ToolGroup = { kind: ToolKind; label: string; calls: ToolCall[] };

export function groupTools(tools: ToolCall[]): ToolGroup[] {
  const groups = new Map<ToolKind, ToolGroup>();
  for (const call of tools) {
    const kind = toolKind(call.name);
    const group = groups.get(kind) ?? {
      kind,
      label: kindLabel(kind),
      calls: [],
    };
    group.calls.push(call);
    groups.set(kind, group);
  }
  return [...groups.values()];
}

const SUMMARY_KEYS = [
  "command",
  "cmd",
  "path",
  "file_path",
  "url",
  "query",
  "pattern",
];

export function summarizeArgs(args: unknown, max = 90): string {
  let text = "";
  if (typeof args === "string") text = args;
  else if (args && typeof args === "object") {
    const record = args as Record<string, unknown>;
    const key = SUMMARY_KEYS.find((k) => typeof record[k] === "string");
    text = key ? String(record[key]) : JSON.stringify(args);
  } else if (args !== undefined && args !== null) text = String(args);
  text = text.replace(/\s+/g, " ").trim();
  return text.length > max ? `${text.slice(0, max - 1)}…` : text;
}

export function prettyArgs(args: unknown): string {
  if (typeof args === "string") return args;
  if (args === undefined || args === null) return "";
  return JSON.stringify(args, null, 2);
}

export function runningCount(group: ToolGroup): number {
  return group.calls.filter((c) => c.status === "running").length;
}

export function failedCount(group: ToolGroup): number {
  return group.calls.filter((c) => c.status === "failed").length;
}

export function approvalLabel(tool: string): string {
  return APPROVAL_TOOLS.has(tool) ? t(`approval.tool.${tool}` as AppKey) : tool;
}

export function approvalIsDestructive(tool: string, args: unknown): boolean {
  const action =
    args && typeof args === "object"
      ? (args as Record<string, unknown>).action
      : undefined;
  return /_delete_/.test(tool) || action === "delete";
}

export function approvalCanRemember(tool: string, args?: unknown): boolean {
  if (approvalIsDestructive(tool, args)) return false;
  return !/_delete|^browser_|web_fetch|files_save|^courrier_(send|save_draft)$/.test(tool);
}

export function approvalSummary(args: unknown): string {
  const fields = approvalFields(args);
  if (!fields || !fields.length) return "";
  return fields
    .slice(0, 3)
    .map((f) => f.value)
    .join(" / ");
}

export type ApprovalField = { key: string; value: string };

export function approvalFields(args: unknown): ApprovalField[] | null {
  if (!args || typeof args !== "object" || Array.isArray(args)) return null;
  const fields: ApprovalField[] = [];
  for (const [key, value] of Object.entries(args as Record<string, unknown>)) {
    if (value === null || value === undefined || value === "") continue;
    if (typeof value === "object" && !Array.isArray(value)) return null;
    const text = Array.isArray(value)
      ? value
          .map((v) => (typeof v === "object" ? JSON.stringify(v) : String(v)))
          .join(", ")
      : String(value);
    fields.push({ key: key.replace(/_/g, " "), value: text });
  }
  return fields;
}

function phrase(name: string, n: number): string {
  if (PHRASE_TOOLS.has(name)) return t(`tools.phrase.${name}` as AppKey, { count: n });
  return t("tools.phrase.other", { label: name.replace(/_/g, " "), count: n });
}

/** One sentence for a message's tool calls, e.g. "Searched mail 3 times, read 2 emails". */
export function summarizeTools(tools: ToolCall[], max = 3): string {
  const counts = new Map<string, number>();
  for (const call of tools)
    counts.set(call.name, (counts.get(call.name) ?? 0) + 1);
  const parts = [...counts].map(([name, n]) => phrase(name, n));
  const shown = parts.slice(0, max);
  if (parts.length > max) shown.push(t("tools.more", { count: parts.length - max }));
  const text = shown.join(", ");
  return text.charAt(0).toLocaleUpperCase(locale()) + text.slice(1);
}
