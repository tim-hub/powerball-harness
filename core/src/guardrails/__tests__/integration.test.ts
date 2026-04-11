/**
 * core/src/guardrails/__tests__/integration.test.ts
 * Harness v3 Guardrail E2E Integration Tests
 *
 * Validates that 9 guard rules work correctly together through
 * the actual hook invocation flow (evaluatePreTool → evaluatePostTool).
 *
 * Difference from unit tests:
 *   - rules.test.ts: Unit tests for individual rule functions
 *   - integration.test.ts: Full-flow tests for PreToolUse → PostToolUse pipeline
 */

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { evaluatePreTool } from "../pre-tool.js";
import { evaluatePostTool } from "../post-tool.js";
import type { HookInput } from "../../types.js";

// ============================================================
// Test helpers
// ============================================================

function buildInput(
  toolName: string,
  toolInput: Record<string, unknown>,
  overrides?: Partial<HookInput>
): HookInput {
  return {
    tool_name: toolName,
    tool_input: toolInput,
    session_id: "test-session",
    cwd: "/test/project",
    plugin_root: "/test/plugin",
    ...overrides,
  };
}

// ============================================================
// PreToolUse → decision integration tests
// ============================================================

describe("E2E: PreToolUse flow", () => {
  // Reset environment variables for each test
  beforeEach(() => {
    delete process.env["HARNESS_WORK_MODE"];
    delete process.env["HARNESS_CODEX_MODE"];
    delete process.env["HARNESS_BREEZING_ROLE"];
    delete process.env["HARNESS_PROJECT_ROOT"];
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  // ------------------------------------------------------------------
  // Happy path: safe operations are approved
  // ------------------------------------------------------------------

  describe("approve cases", () => {
    it("normal Bash command is approved", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "npm test" })
      );
      expect(result.decision).toBe("approve");
    });

    it("normal file write is approved", async () => {
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/src/index.ts" })
      );
      expect(result.decision).toBe("approve");
    });

    it("rm without -r flag is approved", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "rm /tmp/test.log" })
      );
      expect(result.decision).toBe("approve");
    });

    it("git push (without force) is approved", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push origin feature/login" })
      );
      expect(result.decision).toBe("approve");
    });

    it("normal Read is approved", async () => {
      const result = await evaluatePreTool(
        buildInput("Read", { file_path: "/test/project/src/app.ts" })
      );
      expect(result.decision).toBe("approve");
    });
  });

  // ------------------------------------------------------------------
  // deny cases: dangerous operations are blocked
  // ------------------------------------------------------------------

  describe("deny cases", () => {
    it("sudo is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "sudo apt-get update" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("sudo");
    });

    it("git push --force is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push --force origin main" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("force");
    });

    it("git push -f is also denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push -f origin main" })
      );
      expect(result.decision).toBe("deny");
    });

    it("--no-verify is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git commit --no-verify -m 'test'" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("--no-verify");
    });

    it("--no-gpg-sign is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git commit --no-gpg-sign -m 'test'" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("--no-gpg-sign");
    });

    it("git reset --hard main is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git reset --hard main" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("reset --hard");
    });

    it("Write to .env file is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/.env" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain(".env");
    });

    it("Edit to .git/ directory is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Edit", { file_path: "/test/project/.git/config" })
      );
      expect(result.decision).toBe("deny");
    });

    it("Bash writing to .env is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "echo 'SECRET=123' > .env" })
      );
      expect(result.decision).toBe("deny");
    });

    it("Bash writing to .env variants is also denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "echo 'KEY=val' >> .env.local" })
      );
      expect(result.decision).toBe("deny");
    });

    it("Write to private key file is denied", async () => {
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/root/.ssh/id_rsa" })
      );
      expect(result.decision).toBe("deny");
    });

    it("Write is denied in codex mode", async () => {
      process.env["HARNESS_CODEX_MODE"] = "1";
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/src/app.ts" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("Codex");
    });

    it("breezing reviewer is denied git commit", async () => {
      process.env["HARNESS_BREEZING_ROLE"] = "reviewer";
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git commit -m 'feat: add feature'" })
      );
      expect(result.decision).toBe("deny");
    });

    it("breezing reviewer is denied Write", async () => {
      process.env["HARNESS_BREEZING_ROLE"] = "reviewer";
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/src/app.ts" })
      );
      expect(result.decision).toBe("deny");
    });
  });

  // ------------------------------------------------------------------
  // ask cases: operations requiring confirmation
  // ------------------------------------------------------------------

  describe("ask cases (non-work mode)", () => {
    it("rm -rf returns ask", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "rm -rf /tmp/test-dir" })
      );
      expect(result.decision).toBe("ask");
      expect(result.reason).toContain("rm");
    });

    it("rm -fr also returns ask", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "rm -fr dist/" })
      );
      expect(result.decision).toBe("ask");
    });

    it("Write to absolute path outside project returns ask", async () => {
      process.env["HARNESS_PROJECT_ROOT"] = "/test/project";
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/etc/hosts" })
      );
      expect(result.decision).toBe("ask");
    });
  });

  // ------------------------------------------------------------------
  // work mode: bypassable operations
  // ------------------------------------------------------------------

  describe("work mode bypass cases", () => {
    it("rm -rf is not asked in work mode", async () => {
      process.env["HARNESS_WORK_MODE"] = "1";
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "rm -rf dist/" })
      );
      // In work mode, ask is skipped and becomes approve
      expect(result.decision).toBe("approve");
    });

    it("writing outside project is not asked in work mode", async () => {
      process.env["HARNESS_WORK_MODE"] = "1";
      process.env["HARNESS_PROJECT_ROOT"] = "/test/project";
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/tmp/output.txt" })
      );
      expect(result.decision).toBe("approve");
    });

    it("sudo is still denied in work mode (no exceptions)", async () => {
      process.env["HARNESS_WORK_MODE"] = "1";
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "sudo make install" })
      );
      expect(result.decision).toBe("deny");
    });

    it("git push --force is still denied in work mode (no exceptions)", async () => {
      process.env["HARNESS_WORK_MODE"] = "1";
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push --force origin main" })
      );
      expect(result.decision).toBe("deny");
    });
  });

  // ------------------------------------------------------------------
  // approve + systemMessage cases: approved with warning
  // ------------------------------------------------------------------

  describe("approve + warning cases", () => {
    it("Read of .env file is approved but emits a warning", async () => {
      const result = await evaluatePreTool(
        buildInput("Read", { file_path: "/test/project/.env" })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toContain("Warning");
      expect(result.systemMessage).toContain(".env");
    });

    it("git push origin main returns approve + systemMessage", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push origin main" })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeTruthy();
      expect(result.systemMessage).toContain("main");
    });

    it("Write to package.json returns approve + systemMessage", async () => {
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/package.json" })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeTruthy();
      expect(result.systemMessage).toContain("package.json");
    });
  });
});

// ============================================================
// PostToolUse flow
// ============================================================

describe("E2E: PostToolUse flow", () => {
  beforeEach(() => {
    delete process.env["HARNESS_WORK_MODE"];
  });

  it("normal Write result is approved", async () => {
    const result = await evaluatePostTool(
      buildInput("Write", {
        file_path: "/test/project/src/app.ts",
        content: "export const app = {};",
      })
    );
    expect(result.decision).toBe("approve");
  });

  it("test tampering detection returns approve + warning", async () => {
    const result = await evaluatePostTool(
      buildInput("Write", {
        file_path: "/test/project/src/__tests__/app.test.ts",
        content: "it.skip('should work', () => { expect(true).toBe(true); });",
      })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeTruthy();
    expect(result.systemMessage).toContain("it.skip");
  });

  it("ESLint disable comment triggers a warning", async () => {
    const result = await evaluatePostTool(
      buildInput("Write", {
        file_path: "/test/project/.eslintrc.js",
        content: "/* eslint-disable */\nmodule.exports = {};",
      })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeTruthy();
  });

  it("CI continue-on-error addition triggers a warning", async () => {
    const result = await evaluatePostTool(
      buildInput("Write", {
        file_path: "/test/project/.github/workflows/ci.yml",
        content: "continue-on-error: true",
      })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeTruthy();
  });

  it("normal Bash execution result is approved", async () => {
    const result = await evaluatePostTool(
      buildInput("Bash", {
        command: "npm test",
        output: "Tests passed: 42",
      })
    );
    expect(result.decision).toBe("approve");
  });
});

// ============================================================
// Pre → Post sequential flow
// ============================================================

describe("E2E: PreToolUse → PostToolUse sequential flow", () => {
  it("normal Write operation is approved by both hooks", async () => {
    const input = buildInput("Write", {
      file_path: "/test/project/src/feature.ts",
      content: "export const feature = () => 'hello';",
    });

    const preResult = await evaluatePreTool(input);
    expect(preResult.decision).toBe("approve");

    const postResult = await evaluatePostTool(input);
    expect(postResult.decision).toBe("approve");
  });

  it("sudo is blocked at PreToolUse (PostToolUse is unnecessary)", async () => {
    const input = buildInput("Bash", { command: "sudo apt-get install curl" });

    const preResult = await evaluatePreTool(input);
    expect(preResult.decision).toBe("deny");
    // When denied, PostToolUse is not executed (simulation only here)
    expect(preResult.reason).toContain("sudo");
  });
});
