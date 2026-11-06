import { describe, expect, it } from "vitest";
import { groupApprovals } from "./batch";
import type { Approval } from "./reducer";

const make = (callId: string, tool: string, state: Approval["state"] = "pending", args: unknown = {}): Approval => ({
  callId,
  tool,
  args,
  state,
});

describe("groupApprovals", () => {
  it("batches open approvals of the same tool at the first position", () => {
    const entries = groupApprovals([
      make("1", "opus_delete_task"),
      make("2", "agenda_delete_event"),
      make("3", "opus_delete_task"),
    ]);
    expect(entries.map((e) => e.kind)).toEqual(["batch", "single"]);
    const batch = entries[0];
    expect(batch.kind === "batch" && batch.approvals.map((a) => a.callId)).toEqual(["1", "3"]);
  });

  it("keeps a lone approval single", () => {
    expect(groupApprovals([make("1", "opus_delete_task")])[0].kind).toBe("single");
  });

  it("leaves settled approvals out of the batch", () => {
    const entries = groupApprovals([
      make("1", "opus_delete_task", "approved"),
      make("2", "opus_delete_task"),
      make("3", "opus_delete_task"),
    ]);
    expect(entries.map((e) => e.kind)).toEqual(["single", "batch"]);
  });

  it("never batches mail", () => {
    const mail = { kind: "mail", to: ["a@x"] };
    const entries = groupApprovals([
      make("1", "courrier_send", "pending", mail),
      make("2", "courrier_send", "pending", mail),
    ]);
    expect(entries.every((e) => e.kind === "single")).toBe(true);
  });
});

describe('approvalCanRemember', () => {
	it('refuses mail, deletes, files and web tools, allows plain writes', async () => {
		const { approvalCanRemember } = await import('./tools');
		for (const tool of ['courrier_send', 'courrier_save_draft', 'opus_delete_task', 'files_save', 'web_fetch', 'browser_navigate']) {
			expect(approvalCanRemember(tool)).toBe(false);
		}
		for (const tool of ['opus_create_task', 'opus_update_task', 'agenda_create_event']) {
			expect(approvalCanRemember(tool)).toBe(true);
		}
	});
});
