/**
 * core/src/guardrails/__tests__/permission.test.ts
 * Unit tests for permission.ts
 *
 * Verifies that all logic from permission-request.sh is correctly ported.
 */

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { existsSync, readFileSync } from "node:fs";
import { evaluatePermission, formatPermissionOutput } from "../permission.js";
import type { HookInput } from "../../types.js";

// Mock the fs module
vi.mock("node:fs");

const mockedExistsSync = vi.mocked(existsSync);
const mockedReadFileSync = vi.mocked(readFileSync);

function makeInput(
  toolName: string,
  toolInput: Record<string, unknown> = {},
  cwd = "/project"
): HookInput {
  const input: HookInput = { tool_name: toolName, tool_input: toolInput };
  input.cwd = cwd;
  return input;
}

beforeEach(() => {
  vi.clearAllMocks();
  // Default: allowlist file does not exist
  mockedExistsSync.mockReturnValue(false);
});

afterEach(() => {
  vi.restoreAllMocks();
});

// ============================================================
// Edit / Write auto-approval
// ============================================================
describe("Edit / Write auto-approval", () => {
  it("auto-approves Edit tool", () => {
    const result = evaluatePermission(
      makeInput("Edit", { file_path: "/project/src/foo.ts" })
    );
    // Returns approve with PermissionRequest JSON in systemMessage
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeDefined();
    const parsed = JSON.parse(result.systemMessage!);
    expect(parsed.hookSpecificOutput.decision.behavior).toBe("allow");
  });

  it("auto-approves Write tool", () => {
    const result = evaluatePermission(
      makeInput("Write", { file_path: "/project/src/bar.ts", content: "" })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeDefined();
    const parsed = JSON.parse(result.systemMessage!);
    expect(parsed.hookSpecificOutput.decision.behavior).toBe("allow");
  });

  it("auto-approves MultiEdit tool", () => {
    const result = evaluatePermission(makeInput("MultiEdit"));
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeDefined();
  });
});

// ============================================================
// Bash: read-only git command auto-approval
// ============================================================
describe("Bash: read-only git command auto-approval", () => {
  const readOnlyGitCmds = [
    "git status",
    "git diff",
    "git log --oneline -5",
    "git branch -a",
    "git rev-parse HEAD",
    "git show HEAD",
    "git ls-files",
  ];

  for (const cmd of readOnlyGitCmds) {
    it(`auto-approves ${cmd}`, () => {
      const result = evaluatePermission(
        makeInput("Bash", { command: cmd })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeDefined();
      const parsed = JSON.parse(result.systemMessage!);
      expect(parsed.hookSpecificOutput.decision.behavior).toBe("allow");
    });
  }
});

// ============================================================
// Bash: npm/pnpm/yarn — no allowlist means no auto-approval
// ============================================================
describe("Bash: npm/pnpm/yarn — no allowlist means no auto-approval", () => {
  it("npm test defaults to approve without auto-approval when no allowlist", () => {
    mockedExistsSync.mockReturnValue(false);
    const result = evaluatePermission(
      makeInput("Bash", { command: "npm test" })
    );
    // Not safe, so approve without systemMessage
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeUndefined();
  });
});

// ============================================================
// Bash: npm/pnpm/yarn — auto-approved with allowlist
// ============================================================
describe("Bash: npm/pnpm/yarn — auto-approved with allowlist", () => {
  beforeEach(() => {
    mockedExistsSync.mockReturnValue(true);
    mockedReadFileSync.mockReturnValue(
      JSON.stringify({ allowed: true }) as unknown as Buffer
    );
  });

  const pkgCmds = [
    "npm test",
    "npm run test",
    "npm run lint",
    "npm run typecheck",
    "npm run build",
    "npm run validate",
    "pnpm test",
    "yarn test",
    "yarn lint",
  ];

  for (const cmd of pkgCmds) {
    it(`auto-approves ${cmd} (with allowlist)`, () => {
      const result = evaluatePermission(
        makeInput("Bash", { command: cmd })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeDefined();
      const parsed = JSON.parse(result.systemMessage!);
      expect(parsed.hookSpecificOutput.decision.behavior).toBe("allow");
    });
  }
});

// ============================================================
// Bash: Python / Go / Rust test auto-approval
// ============================================================
describe("Bash: Python / Go / Rust test auto-approval", () => {
  const safeCmds = [
    "pytest",
    "pytest -v",
    "python -m pytest",
    "go test ./...",
    "cargo test",
  ];

  for (const cmd of safeCmds) {
    it(`auto-approves ${cmd}`, () => {
      const result = evaluatePermission(
        makeInput("Bash", { command: cmd })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeDefined();
    });
  }
});

// ============================================================
// Bash: commands with shell special characters are not auto-approved
// ============================================================
describe("Bash: commands with shell special characters are not auto-approved", () => {
  const dangerousCmds = [
    "git status | grep modified",
    "npm test && git push",
    "npm test; echo done",
    "echo $HOME",
    "cat file > output",
    "cmd `dangerous`",
  ];

  for (const cmd of dangerousCmds) {
    it(`"${cmd}" defaults to approve only`, () => {
      const result = evaluatePermission(
        makeInput("Bash", { command: cmd })
      );
      // Conservative: returns approve but without systemMessage (no auto-approval)
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeUndefined();
    });
  }
});

// ============================================================
// Other tools: default behavior
// ============================================================
describe("other tools default to approve", () => {
  const otherTools = ["Read", "Glob", "Grep", "Task", "Skill"];

  for (const tool of otherTools) {
    it(`${tool} defaults to approve`, () => {
      const result = evaluatePermission(makeInput(tool));
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeUndefined();
    });
  }
});

// ============================================================
// formatPermissionOutput
// ============================================================
describe("formatPermissionOutput", () => {
  it("outputs correct systemMessage containing PermissionResponse JSON", () => {
    const result = evaluatePermission(
      makeInput("Edit", { file_path: "/project/src/foo.ts" })
    );
    const output = formatPermissionOutput(result);
    const parsed = JSON.parse(output);
    expect(parsed.hookSpecificOutput.hookEventName).toBe("PermissionRequest");
    expect(parsed.hookSpecificOutput.decision.behavior).toBe("allow");
  });

  it("outputs normal HookResult when no systemMessage", () => {
    const result = evaluatePermission(makeInput("Bash", { command: "rm -rf /" }));
    const output = formatPermissionOutput(result);
    const parsed = JSON.parse(output);
    expect(parsed.decision).toBe("approve");
  });
});
