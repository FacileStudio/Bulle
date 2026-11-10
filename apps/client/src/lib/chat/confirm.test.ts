import { describe, expect, it } from "vitest";
import { serverQuestion } from "./confirm";

describe("serverQuestion", () => {
  it("reads a server's confirmation and ignores every other card", () => {
    const question = serverQuestion({
      kind: "mcp",
      server: "Courrier",
      message: 'Send the email "Hi" to ann@x now?',
      arguments: { to: ["ann@x"], body: "Hi" },
    });
    expect(question?.server).toBe("Courrier");
    expect(question?.arguments.body).toBe("Hi");
    expect(serverQuestion({ kind: "mail", to: ["a@x"] })).toBeNull();
    expect(serverQuestion("raw")).toBeNull();
  });
});
