/**
 * core/src/__tests__/types.test.ts
 * Basic consistency checks for type definitions
 */

import { describe, it, expect } from "vitest";
import type {
  HookInput,
  HookResult,
  GuardRule,
  Signal,
  TaskFailure,
  SessionState,
} from "../types.js";

describe("HookInput", () => {
  it("can be constructed with minimal fields", () => {
    const input: HookInput = {
      tool_name: "Bash",
      tool_input: { command: "ls" },
    };
    expect(input.tool_name).toBe("Bash");
    expect(input.tool_input).toEqual({ command: "ls" });
  });

  it("can be constructed with optional fields", () => {
    const input: HookInput = {
      tool_name: "Write",
      tool_input: { file_path: "/tmp/test.ts", content: "" },
      session_id: "sess-123",
      cwd: "/project",
      plugin_root: "/plugin",
    };
    expect(input.session_id).toBe("sess-123");
    expect(input.cwd).toBe("/project");
    expect(input.plugin_root).toBe("/plugin");
  });
});

describe("HookResult", () => {
  it("can represent an approve decision", () => {
    const result: HookResult = { decision: "approve" };
    expect(result.decision).toBe("approve");
  });

  it("can represent a deny decision with reason", () => {
    const result: HookResult = {
      decision: "deny",
      reason: "Protected path",
      systemMessage: "Cannot write to .git/",
    };
    expect(result.decision).toBe("deny");
    expect(result.reason).toBe("Protected path");
    expect(result.systemMessage).toBe("Cannot write to .git/");
  });

  it("can represent an ask decision", () => {
    const result: HookResult = {
      decision: "ask",
      reason: "Confirm git push?",
    };
    expect(result.decision).toBe("ask");
  });
});

describe("GuardRule", () => {
  it("can be constructed with correct structure", () => {
    const rule: GuardRule = {
      id: "block-git-dir",
      toolPattern: /^(Write|Edit)$/,
      evaluate: (ctx) => {
        const path = ctx.input.tool_input["file_path"];
        if (typeof path === "string" && path.includes(".git/")) {
          return { decision: "deny", reason: "Protected .git/ directory" };
        }
        return null;
      },
    };

    expect(rule.id).toBe("block-git-dir");
    expect(rule.toolPattern.test("Write")).toBe(true);
    expect(rule.toolPattern.test("Bash")).toBe(false);

    const mockCtx = {
      input: {
        tool_name: "Write",
        tool_input: { file_path: "/project/.git/config" },
      },
      projectRoot: "/project",
      workMode: false,
      codexMode: false,
      breezingRole: null,
    };

    const result = rule.evaluate(mockCtx);
    expect(result).not.toBeNull();
    expect(result?.decision).toBe("deny");
  });

  it("returns null when no match", () => {
    const rule: GuardRule = {
      id: "test-rule",
      toolPattern: /^Bash$/,
      evaluate: () => null,
    };
    expect(rule.evaluate({
      input: { tool_name: "Bash", tool_input: {} },
      projectRoot: "/project",
      workMode: false,
      codexMode: false,
      breezingRole: null,
    })).toBeNull();
  });
});

describe("Signal", () => {
  it("can construct a signal", () => {
    const signal: Signal = {
      type: "task_completed",
      from_session_id: "sess-abc",
      payload: { task_id: "task-1", status: "success" },
      timestamp: new Date().toISOString(),
    };
    expect(signal.type).toBe("task_completed");
    expect(signal.from_session_id).toBe("sess-abc");
    expect(signal.to_session_id).toBeUndefined();
  });
});

describe("TaskFailure", () => {
  it("can construct a task failure event", () => {
    const failure: TaskFailure = {
      task_id: "task-1",
      severity: "error",
      message: "Build failed",
      timestamp: new Date().toISOString(),
      attempt: 1,
    };
    expect(failure.severity).toBe("error");
    expect(failure.attempt).toBe(1);
  });
});

describe("SessionState", () => {
  it("can construct a session state", () => {
    const state: SessionState = {
      session_id: "sess-xyz",
      mode: "work",
      project_root: "/project",
      started_at: new Date().toISOString(),
    };
    expect(state.mode).toBe("work");
  });
});
