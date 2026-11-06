import type { ToolCall } from "./reducer";

export type ToolKind = "apps" | "commands" | "files" | "web" | "other";

export const KIND_LABEL: Record<ToolKind, string> = {
  apps: "Facile apps",
  commands: "Commands",
  files: "Files",
  web: "Web",
  other: "Other tools",
};

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
      label: KIND_LABEL[kind],
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

const APPROVAL_LABEL: Record<string, string> = {
  courrier_send: "Send an email",
  courrier_save_draft: "Save a draft",
  courrier_delete_draft: "Delete a draft",
  courrier_update: "Change emails",
  courrier_move: "Move emails",
  agenda_create_event: "Create an event",
  agenda_update_event: "Change an event",
  agenda_delete_event: "Delete an event",
  opus_create_task: "Create a task",
  opus_update_task: "Change a task",
  opus_delete_task: "Delete a task",
  sablier_start: "Start a timer",
  sablier_stop: "Stop the timer",
  glouton_save_draft: "Save a prospect draft",
  glouton_update_draft: "Change a prospect draft",
  glouton_suggest: "Suggest a prospect action",
  glouton_create_company: "Create a company",
  glouton_update_company: "Change a company",
  glouton_create_person: "Create a contact",
  glouton_update_person: "Change a contact",
  browser_confirm: "Send a web request",
};

export function approvalLabel(tool: string): string {
  return APPROVAL_LABEL[tool] ?? tool;
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

const TOOL_PHRASE: Record<string, [string, string]> = {
  courrier_inbox: ["checked the inbox", "checked the inbox {n} times"],
  courrier_search: ["searched mail", "searched mail {n} times"],
  courrier_read: ["read an email", "read {n} emails"],
  courrier_accounts: ["listed mailboxes", "listed mailboxes"],
  courrier_save_draft: ["saved a draft", "saved {n} drafts"],
  courrier_send: ["sent an email", "sent {n} emails"],
  agenda_events: ["checked the calendar", "checked the calendar {n} times"],
  agenda_create_event: ["created an event", "created {n} events"],
  opus_tasks: ["checked tasks", "checked tasks {n} times"],
  opus_create_task: ["created a task", "created {n} tasks"],
  glouton_contacts: ["looked up contacts", "looked up contacts {n} times"],
  memory_search: ["searched memory", "searched memory {n} times"],
  memory_write: ["saved a note", "saved {n} notes"],
  web_search: ["searched the web", "searched the web {n} times"],
  files_save: ["saved a file", "saved {n} files"],
};

function phrase(name: string, n: number): string {
  const known = TOOL_PHRASE[name];
  if (known) return (n === 1 ? known[0] : known[1]).replace("{n}", String(n));
  const label = name.replace(/_/g, " ");
  return n === 1 ? `used ${label}` : `used ${label} ${n} times`;
}

/** One sentence for a message's tool calls, e.g. "Searched mail 3 times, read 2 emails". */
export function summarizeTools(tools: ToolCall[], max = 3): string {
  const counts = new Map<string, number>();
  for (const call of tools)
    counts.set(call.name, (counts.get(call.name) ?? 0) + 1);
  const parts = [...counts].map(([name, n]) => phrase(name, n));
  const shown = parts.slice(0, max);
  if (parts.length > max) shown.push(`${parts.length - max} more`);
  const text = shown.join(", ");
  return text.charAt(0).toUpperCase() + text.slice(1);
}
