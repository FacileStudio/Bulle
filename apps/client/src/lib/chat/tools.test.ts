import { describe, expect, it } from "vitest";
import type { ToolCall } from "./reducer";
import { summarizeTools } from "./tools";

const call = (name: string): ToolCall =>
  ({ callId: name + Math.random(), name, args: {}, status: "ok" }) as ToolCall;

describe("summarizeTools", () => {
  it("counts repeated tools into one phrase", () => {
    const tools = [
      "courrier_search",
      "courrier_search",
      "courrier_read",
      "courrier_read",
      "courrier_save_draft",
    ].map(call);
    expect(summarizeTools(tools)).toBe(
      "Searched mail 2 times, read 2 emails, saved a draft",
    );
  });

  it("folds the tail and names unknown tools", () => {
    const tools = ["a_b", "courrier_inbox", "web_search", "memory_search"].map(
      call,
    );
    expect(summarizeTools(tools)).toBe(
      "Used a b, checked the inbox, searched the web, 1 more",
    );
  });
});
