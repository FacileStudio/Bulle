import { describe, expect, it } from "vitest";
import { draftOf, mailEdits, mailRecord } from "./mail";

const args = {
  kind: "mail",
  from: "me@ik.me",
  to: ["lisa1.nay@epitech.eu"],
  subject: "Re: Visio de point",
  body: "Bonjour",
  threaded: true,
};

describe("mail approval", () => {
  it("reads only mail records", () => {
    expect(mailRecord({ to: "x" })).toBeNull();
    expect(mailRecord(args)?.threaded).toBe(true);
  });

  it("sends nothing back when the owner changed nothing", () => {
    const mail = mailRecord(args)!;
    expect(mailEdits(mail, draftOf(mail))).toBeUndefined();
  });

  it("returns only the changed fields, addresses split", () => {
    const mail = mailRecord(args)!;
    const draft = {
      ...draftOf(mail),
      body: "Bonjour Madame Nay",
      cc: "a@x; b@x",
    };
    expect(mailEdits(mail, draft)).toEqual({
      body: "Bonjour Madame Nay",
      cc: ["a@x", "b@x"],
    });
  });
});
