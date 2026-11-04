/** The email an approval card shows, as the API's mail policy describes it. */
export type MailRecord = {
  from: string;
  to: string[];
  cc: string[];
  bcc: string[];
  subject: string;
  body: string;
  threaded: boolean;
};

/** Editable copy of a mail record, addresses as comma-separated text. */
export type MailDraft = {
  to: string;
  cc: string;
  bcc: string;
  subject: string;
  body: string;
};

function list(value: unknown): string[] {
  if (Array.isArray(value)) return value.map(String).filter(Boolean);
  if (typeof value === "string" && value) return [value];
  return [];
}

/** Reads a mail record from approval args, or null when they are not one. */
export function mailRecord(args: unknown): MailRecord | null {
  if (!args || typeof args !== "object" || Array.isArray(args)) return null;
  const record = args as Record<string, unknown>;
  if (record.kind !== "mail") return null;
  return {
    from: typeof record.from === "string" ? record.from : "",
    to: list(record.to),
    cc: list(record.cc),
    bcc: list(record.bcc),
    subject: typeof record.subject === "string" ? record.subject : "",
    body: typeof record.body === "string" ? record.body : "",
    threaded: record.threaded === true,
  };
}

export function draftOf(mail: MailRecord): MailDraft {
  return {
    to: mail.to.join(", "),
    cc: mail.cc.join(", "),
    bcc: mail.bcc.join(", "),
    subject: mail.subject,
    body: mail.body,
  };
}

function addresses(text: string): string[] {
  return text
    .split(/[,;\n]/)
    .map((part) => part.trim())
    .filter(Boolean);
}

/** The fields the owner changed, in the shape courrier_send takes; undefined when nothing changed. */
export function mailEdits(
  mail: MailRecord,
  draft: MailDraft,
): Record<string, unknown> | undefined {
  const edits: Record<string, unknown> = {};
  for (const key of ["to", "cc", "bcc"] as const) {
    const next = addresses(draft[key]);
    if (next.join(",") !== mail[key].join(",")) edits[key] = next;
  }
  if (draft.subject !== mail.subject) edits.subject = draft.subject;
  if (draft.body !== mail.body) edits.body = draft.body;
  return Object.keys(edits).length ? edits : undefined;
}
