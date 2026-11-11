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
      "courrier_search_mail",
      "courrier_search_mail",
      "courrier_read_thread",
      "courrier_read_thread",
      "courrier_draft",
    ].map(call);
    expect(summarizeTools(tools)).toBe(
      "Searched mail 2 times, read 2 conversations, saved a draft",
    );
  });

  it("folds the tail and names unknown tools", () => {
    const tools = [
      "a_b",
      "courrier_organize",
      "sonar_search",
      "memory_search",
    ].map(call);
    expect(summarizeTools(tools)).toBe(
      "Used a b, sorted emails, searched the web, 1 more",
    );
  });

  it("phrases the renamed suite and job tools", () => {
    const tools = ["nuage_read_file", "opus_search_tasks", "jobs_create"].map(call);
    expect(summarizeTools(tools)).toBe(
      "Read a file, checked tasks, scheduled a job",
    );
  });

  it("phrases the nacelle v0.32 tool names", () => {
    const tools = ["bash", "bash", "grep", "glob"].map(call);
    expect(summarizeTools(tools, 4)).toBe("Ran 2 commands, searched the files, listed files");
    expect(summarizeTools(["todo_write"].map(call))).toBe("Updated the plan");
  });
});

describe("toolKind", () => {
  it("counts nuage and jobs as apps, memory and chat search as other", () => {
    expect(toolKind("nuage_search")).toBe("apps");
    expect(toolKind("jobs_list")).toBe("apps");
    expect(toolKind("glouton_get_person")).toBe("apps");
    expect(toolKind("vision_get_stats")).toBe("apps");
    expect(toolKind("echo_read_transcript")).toBe("apps");
    expect(toolKind("memory_read")).toBe("other");
    expect(toolKind("chats_search")).toBe("other");
  });
});

describe("approvals", () => {
  it("labels the new gated tools and never remembers a send, a trash or a public link", () => {
    expect(approvalLabel("jobs_create")).toBe("Schedule a job");
    expect(approvalLabel("nuage_create_share")).toBe("Create a public link");
    expect(approvalCanRemember("courrier_send")).toBe(false);
    expect(approvalCanRemember("nuage_delete")).toBe(false);
    expect(approvalCanRemember("nuage_create_share")).toBe(false);
    expect(approvalCanRemember("opus_update_task")).toBe(true);
    expect(approvalLabel("capsule_revoke")).toBe("Destroy a capsule");
    expect(approvalCanRemember("plume_send_document")).toBe(false);
    expect(approvalCanRemember("vision_share_site")).toBe(false);
  });
});
