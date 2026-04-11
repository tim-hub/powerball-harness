/**
 * core/src/guardrails/post-tool.ts
 * PostToolUse hook integrated evaluation function
 *
 * Runs the following PostToolUse script equivalents in parallel via Promise.allSettled
 * and aggregates the results into a single HookResult:
 *
 * 1. tampering-detector: Test tampering detection (warning only)
 * 2. security-review: Security pattern detection (warning only)
 *
 * Others (log-toolname, commit-cleanup, etc.) are side-effect-only and do not affect
 * the HookResult, so they maintain their design as separate hooks.json entries.
 */

import type { HookInput, HookResult } from "../types.js";
import { detectTestTampering } from "./tampering.js";

// ============================================================
// Security pattern detection (ported from posttooluse-security-review.sh)
// ============================================================

/**
 * Detect security risk patterns in written code.
 * If detected, add warnings as systemMessage (does not block).
 */
function detectSecurityRisks(input: HookInput): string[] {
  const toolInput = input.tool_input;
  const content =
    typeof toolInput["content"] === "string"
      ? toolInput["content"]
      : typeof toolInput["new_string"] === "string"
        ? toolInput["new_string"]
        : null;

  if (content === null) return [];

  const warnings: string[] = [];

  const securityPatterns: Array<{ pattern: RegExp; message: string }> = [
    {
      pattern: /process\.env\.[A-Z_]+.*(?:password|secret|key|token)/i,
      message: "Possible sensitive data embedded directly from environment variables",
    },
    {
      pattern: /eval\s*\(\s*(?:request|req|input|param|query)/i,
      message: "Detected code passing user input to eval() (RCE risk)",
    },
    {
      pattern: /exec\s*\(\s*`[^`]*\$\{/,
      message: "Detected code passing template literals to exec() (command injection risk)",
    },
    {
      pattern: /innerHTML\s*=\s*(?:.*\+.*|`[^`]*\$\{)/,
      message: "Detected code setting user input to innerHTML (XSS risk)",
    },
    {
      pattern: /(?:password|passwd|secret|api_key|apikey)\s*=\s*["'][^"']{8,}["']/i,
      message: "Detected hardcoded sensitive data (password/API key)",
    },
  ];

  for (const { pattern, message } of securityPatterns) {
    if (pattern.test(content)) {
      warnings.push(message);
    }
  }

  return warnings;
}

// ============================================================
// PostToolUse integrated entry point
// ============================================================

/**
 * PostToolUse hook entry point.
 * Runs multiple detectors in parallel and returns aggregated warnings.
 */
export async function evaluatePostTool(input: HookInput): Promise<HookResult> {
  // Only perform detailed checks for Write / Edit / MultiEdit
  const isWriteOp = ["Write", "Edit", "MultiEdit"].includes(input.tool_name);

  if (!isWriteOp) {
    return { decision: "approve" };
  }

  // Parallel execution (Promise.allSettled ensures one failure doesn't affect the other)
  const [tamperingResult, securityWarnings] = await Promise.allSettled([
    Promise.resolve(detectTestTampering(input)),
    Promise.resolve(detectSecurityRisks(input)),
  ]);

  const systemMessages: string[] = [];

  // Collect tampering detection warnings
  if (
    tamperingResult.status === "fulfilled" &&
    tamperingResult.value.systemMessage
  ) {
    systemMessages.push(tamperingResult.value.systemMessage);
  }

  // Collect security warnings
  if (
    securityWarnings.status === "fulfilled" &&
    securityWarnings.value.length > 0
  ) {
    const secLines = securityWarnings.value
      .map((w) => `- ${w}`)
      .join("\n");
    systemMessages.push(
      `[Harness v3] Security risk detected:\n${secLines}`
    );
  }

  if (systemMessages.length === 0) {
    return { decision: "approve" };
  }

  return {
    decision: "approve",
    systemMessage: systemMessages.join("\n\n---\n\n"),
  };
}
