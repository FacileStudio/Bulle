import { serverQuestion } from "./confirm";
import type { Approval } from "./reducer";

export type ApprovalEntry =
  | { kind: "single"; approval: Approval }
  | { kind: "batch"; tool: string; approvals: Approval[] };

function open(approval: Approval): boolean {
  return approval.state === "pending" || approval.state === "sending";
}

export function groupApprovals(approvals: Approval[]): ApprovalEntry[] {
  const byTool = new Map<string, Approval[]>();
  for (const approval of approvals) {
    if (!open(approval) || serverQuestion(approval.args)) continue;
    byTool.set(approval.tool, [...(byTool.get(approval.tool) ?? []), approval]);
  }
  const entries: ApprovalEntry[] = [];
  const placed = new Set<string>();
  for (const approval of approvals) {
    const group = byTool.get(approval.tool);
    if (group && group.length > 1 && group.includes(approval)) {
      if (!placed.has(approval.tool)) {
        placed.add(approval.tool);
        entries.push({ kind: "batch", tool: approval.tool, approvals: group });
      }
      continue;
    }
    entries.push({ kind: "single", approval });
  }
  return entries;
}
