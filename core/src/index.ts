/**
 * core/src/index.ts
 * Harness v3 core engine entry point
 *
 * Reads JSON from stdin, routes by hook type,
 * and writes a JSON response to stdout.
 *
 * Usage:
 *   echo '{"tool_name":"Bash","tool_input":{...}}' | node dist/index.js pre-tool
 *   echo '{"tool_name":"Write","tool_input":{...}}' | node dist/index.js post-tool
 */

import { type HookInput, type HookResult } from "./types.js";

/** Supported hook types */
type HookType = "pre-tool" | "post-tool" | "permission";

/**
 * Read all of stdin and return as a string
 */
async function readStdin(): Promise<string> {
  const chunks: Buffer[] = [];
  for await (const chunk of process.stdin) {
    chunks.push(chunk as Buffer);
  }
  return Buffer.concat(chunks).toString("utf-8");
}

/**
 * Parse stdin JSON into a HookInput
 */
function parseInput(raw: string): HookInput {
  const parsed: unknown = JSON.parse(raw);

  if (
    typeof parsed !== "object" ||
    parsed === null ||
    !("tool_name" in parsed) ||
    typeof (parsed as Record<string, unknown>)["tool_name"] !== "string"
  ) {
    throw new Error("Invalid hook input: missing required field 'tool_name'");
  }

  const obj = parsed as Record<string, unknown>;

  const result: HookInput = {
    tool_name: obj["tool_name"] as string,
    tool_input:
      typeof obj["tool_input"] === "object" && obj["tool_input"] !== null
        ? (obj["tool_input"] as Record<string, unknown>)
        : {},
  };

  if (typeof obj["session_id"] === "string") {
    result.session_id = obj["session_id"];
  }
  if (typeof obj["cwd"] === "string") {
    result.cwd = obj["cwd"];
  }
  if (typeof obj["plugin_root"] === "string") {
    result.plugin_root = obj["plugin_root"];
  }

  return result;
}

/**
 * Route to the appropriate handler based on hook type.
 * Extension point where implementations are added per phase.
 */
async function route(
  hookType: HookType,
  input: HookInput
): Promise<HookResult> {
  switch (hookType) {
    case "pre-tool": {
      const { evaluatePreTool } = await import("./guardrails/pre-tool.js");
      return evaluatePreTool(input);
    }
    case "post-tool": {
      const { evaluatePostTool } = await import("./guardrails/post-tool.js");
      return evaluatePostTool(input);
    }
    case "permission": {
      const { evaluatePermission, formatPermissionOutput } = await import(
        "./guardrails/permission.js"
      );
      const permResult = evaluatePermission(input);
      // PermissionRequest requires hookSpecificOutput format, so we convert
      // via formatPermissionOutput and store the final JSON in systemMessage
      // to bypass main()'s JSON.stringify
      const permJson = formatPermissionOutput(permResult);
      return { decision: permResult.decision, systemMessage: permJson };
    }
    default: {
      // Unknown hook types default to approve (safe fallback)
      return {
        decision: "approve",
        reason: `Unknown hook type: ${String(hookType)}`,
      };
    }
  }
}

/**
 * Convert an error to HookResult format
 */
function errorToResult(err: unknown): HookResult {
  const message = err instanceof Error ? err.message : String(err);
  return {
    decision: "approve",
    reason: `Core engine error (safe fallback): ${message}`,
  };
}

/**
 * Main function: stdin → parse → route → stdout
 */
async function main(): Promise<void> {
  const hookType = (process.argv[2] ?? "pre-tool") as HookType;

  let result: HookResult;

  try {
    const raw = await readStdin();

    if (!raw.trim()) {
      // Empty input safely approves
      result = { decision: "approve", reason: "Empty input" };
    } else {
      const input = parseInput(raw);
      result = await route(hookType, input);
    }
  } catch (err) {
    result = errorToResult(err);
  }

  // For permission hooks, the final JSON is stored in systemMessage
  if (hookType === "permission" && result.systemMessage !== undefined) {
    process.stdout.write(result.systemMessage + "\n");
  } else {
    process.stdout.write(JSON.stringify(result) + "\n");
  }
}

main().catch((err: unknown) => {
  const message = err instanceof Error ? err.message : String(err);
  process.stderr.write(`Fatal: ${message}\n`);
  process.exit(1);
});
