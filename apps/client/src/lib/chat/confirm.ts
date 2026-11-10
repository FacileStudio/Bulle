/** A confirmation an MCP server asked for mid-call, as the API's approval card carries it. */
export type ServerQuestion = {
  server: string;
  message: string;
  url: string;
  arguments: Record<string, unknown>;
};

/** Reads a server's question from approval args, or null when they are not one. */
export function serverQuestion(args: unknown): ServerQuestion | null {
  if (!args || typeof args !== "object" || Array.isArray(args)) return null;
  const record = args as Record<string, unknown>;
  if (record.kind !== "mcp") return null;
  const given = record.arguments;
  return {
    server: typeof record.server === "string" ? record.server : "",
    message: typeof record.message === "string" ? record.message : "",
    url: typeof record.url === "string" ? record.url : "",
    arguments:
      given && typeof given === "object" && !Array.isArray(given)
        ? (given as Record<string, unknown>)
        : {},
  };
}
