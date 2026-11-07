import { describe, expect, it } from "vitest";
import type { ToolCall } from "./reducer";
import {
  approvalCanRemember,
  approvalLabel,
  summarizeTools,
  toolKind,
} from "./tools";

const call = (name: string): ToolCall =>
  ({ callId: name + Math.random(), name, args: {}, status: "ok" }) as ToolCall;

describe("summarizeTools", () => {
  it("counts repeated tools into one phrase", () => {
    const tools = [
      "courrier_search",
      "courrier_search",
      "courrier_get_email",
      "courrier_get_email",
      "courrier_save_draft",
    ].map(call);
    expect(summarizeTools(tools)).toBe(
      "Searched mail 2 times, read 2 emails, saved a draft",
    );
  });

  it("folds the tail and names unknown tools", () => {
    const tools = [
      "a_b",
      "courrier_list_inbox",
      "web_search",
      "memory_search",
    ].map(call);
    expect(summarizeTools(tools)).toBe(
      "Used a b, checked the inbox, searched the web, 1 more",
    );
  });

  it("phrases the renamed suite and job tools", () => {
    const tools = ["nuage_read", "opus_list_tasks", "jobs_create"].map(call);
    expect(summarizeTools(tools)).toBe(
      "Read a file, checked tasks, scheduled a job",
    );
  });
});

describe("toolKind", () => {
  it("counts nuage and jobs as apps, memory and chat search as other", () => {
    expect(toolKind("nuage_search")).toBe("apps");
    expect(toolKind("jobs_list")).toBe("apps");
    expect(toolKind("glouton_get_person")).toBe("apps");
    expect(toolKind("memory_read")).toBe("other");
    expect(toolKind("chats_search")).toBe("other");
  });
});

describe("approvals", () => {
  it("labels the new gated tools and never remembers a move or a file save", () => {
    expect(approvalLabel("jobs_create")).toBe("Schedule a job");
    expect(approvalLabel("nuage_save")).toBe("Save a file");
    expect(approvalCanRemember("courrier_move")).toBe(false);
    expect(approvalCanRemember("nuage_save")).toBe(false);
    expect(approvalCanRemember("opus_update_task")).toBe(true);
  });
});
