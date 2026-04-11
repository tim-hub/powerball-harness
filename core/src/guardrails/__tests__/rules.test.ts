/**
 * core/src/guardrails/__tests__/rules.test.ts
 * Unit tests for GUARD_RULES declarative guard rule table
 *
 * Verifies that each rule from pretooluse-guard.sh is correctly ported to TypeScript.
 * Coverage target: 90%+
 */

import { describe, it, expect } from "vitest";
import { GUARD_RULES, evaluateRules } from "../rules.js";
import type { RuleContext, HookInput } from "../../types.js";

// ============================================================
// Test helpers
// ============================================================

function makeCtx(
  toolName: string,
  toolInput: Record<string, unknown> = {},
  overrides: Partial<Omit<RuleContext, "input">> = {}
): RuleContext {
  const input: HookInput = { tool_name: toolName, tool_input: toolInput };
  return {
    input,
    projectRoot: "/project",
    workMode: false,
    codexMode: false,
    breezingRole: null,
    ...overrides,
  };
}

// ============================================================
// R01: sudo block
// ============================================================
describe("R01: sudo block", () => {
  it("blocks sudo rm -rf /", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "sudo rm -rf /" })
    );
    expect(result.decision).toBe("deny");
  });

  it("blocks sudo apt-get install", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "sudo apt-get install vim" })
    );
    expect(result.decision).toBe("deny");
  });

  it("does not block commands without sudo prefix", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "nosudo echo test" })
    );
    expect(result.decision).toBe("approve");
  });

  it("does not block Bash without sudo", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "ls -la" })
    );
    expect(result.decision).toBe("approve");
  });

  it("does not apply to Write tool", () => {
    // R01 targets Bash only
    const rule = GUARD_RULES.find((r) => r.id === "R01:no-sudo")!;
    const result = rule.evaluate(
      makeCtx("Write", { file_path: "/project/sudo.ts" })
    );
    expect(result).toBeNull();
  });
});

// ============================================================
// R02: protected path write block
// ============================================================
describe("R02: protected path write block", () => {
  const protectedPaths = [
    ".git/config",
    "/project/.git/HEAD",
    ".env",
    "/project/.env",
    "/home/user/.env.local",
    "credentials.pem",
    "private.key",
    "id_rsa",
    "id_ed25519",
    "/home/user/.ssh/id_ecdsa",
  ];

  for (const path of protectedPaths) {
    it(`blocks Write to ${path}`, () => {
      const result = evaluateRules(
        makeCtx("Write", { file_path: path })
      );
      expect(result.decision).toBe("deny");
    });

    it(`blocks Edit to ${path}`, () => {
      const result = evaluateRules(
        makeCtx("Edit", { file_path: path })
      );
      expect(result.decision).toBe("deny");
    });
  }

  it("does not block Write to normal source file", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "/project/src/index.ts" })
    );
    expect(result.decision).toBe("approve");
  });

  it("R02 does not apply to Bash tool", () => {
    const rule = GUARD_RULES.find((r) => r.id === "R02:no-write-protected-paths")!;
    const result = rule.evaluate(
      makeCtx("Bash", { command: "echo hello > .env" })
    );
    expect(result).toBeNull();
  });
});

// ============================================================
// R03: Bash shell write to protected paths block
// ============================================================
describe("R03: Bash shell write to protected paths block", () => {
  const dangerousBashCmds = [
    'echo "SECRET=foo" > .env',
    'echo "key" > .env.local',
    "cat token.txt > .git/config",
    "tee .git/hooks/pre-commit",
    "cat private.key > backup.key",
  ];

  for (const cmd of dangerousBashCmds) {
    it(`blocks ${cmd}`, () => {
      const result = evaluateRules(makeCtx("Bash", { command: cmd }));
      expect(result.decision).toBe("deny");
    });
  }

  it("does not block safe Bash commands", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "echo hello" })
    );
    expect(result.decision).toBe("approve");
  });
});

// ============================================================
// R04: write outside project confirmation
// ============================================================
describe("R04: write outside project confirmation", () => {
  it("returns ask for Write to absolute path outside project", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "/tmp/output.txt" }, { projectRoot: "/project" })
    );
    expect(result.decision).toBe("ask");
  });

  it("returns ask for Edit to absolute path outside project", () => {
    const result = evaluateRules(
      makeCtx("Edit", { file_path: "/home/user/outside.ts" }, { projectRoot: "/project" })
    );
    expect(result.decision).toBe("ask");
  });

  it("does not return ask for absolute path inside project", () => {
    const result = evaluateRules(
      makeCtx(
        "Write",
        { file_path: "/project/src/foo.ts" },
        { projectRoot: "/project" }
      )
    );
    expect(result.decision).toBe("approve");
  });

  it("treats relative paths as inside project", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "src/foo.ts" })
    );
    expect(result.decision).toBe("approve");
  });

  it("does not confirm outside-project writes in work mode", () => {
    const result = evaluateRules(
      makeCtx(
        "Write",
        { file_path: "/tmp/output.txt" },
        { workMode: true, projectRoot: "/project" }
      )
    );
    // R04 is skipped in workMode → subsequent rules approve
    expect(result.decision).toBe("approve");
  });
});

// ============================================================
// R05: rm -rf confirmation
// ============================================================
describe("R05: rm -rf confirmation", () => {
  const rmRfCmds = [
    "rm -rf /tmp/work",
    "rm -fr /tmp/work",
    "rm --recursive /tmp/work",
    "rm -rf ~/Downloads/old",
  ];

  for (const cmd of rmRfCmds) {
    it(`${cmd} returns ask`, () => {
      const result = evaluateRules(makeCtx("Bash", { command: cmd }));
      expect(result.decision).toBe("ask");
    });
  }

  it("does not confirm rm -rf in work mode", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "rm -rf /tmp/work" }, { workMode: true })
    );
    // R05 skipped in workMode → falls through to R06 (no match → approve)
    expect(result.decision).toBe("approve");
  });

  it("does not block normal rm -f", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "rm -f /tmp/test.log" })
    );
    expect(result.decision).toBe("approve");
  });
});

// ============================================================
// R06: git push --force block
// ============================================================
describe("R06: git push --force block", () => {
  const forcePushCmds = [
    "git push --force",
    "git push --force-with-lease",
    "git push origin main --force",
    "git push -f",
    "git push origin main -f",
  ];

  for (const cmd of forcePushCmds) {
    it(`blocks ${cmd}`, () => {
      const result = evaluateRules(makeCtx("Bash", { command: cmd }));
      expect(result.decision).toBe("deny");
    });
  }

  it("does not block normal git push", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "git push origin feature/login" })
    );
    expect(result.decision).toBe("approve");
  });

  it("blocks force push even in work mode (no exceptions)", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "git push --force" }, { workMode: true })
    );
    expect(result.decision).toBe("deny");
  });
});

// ============================================================
// R10: Git bypass flags block
// ============================================================
describe("R10: Git bypass flags block", () => {
  it("blocks --no-verify", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "git commit --no-verify -m 'test'" })
    );
    expect(result.decision).toBe("deny");
  });

  it("blocks --no-gpg-sign", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "git commit --no-gpg-sign -m 'test'" })
    );
    expect(result.decision).toBe("deny");
  });
});

// ============================================================
// R11: git reset --hard to protected branch block
// ============================================================
describe("R11: git reset --hard to protected branch block", () => {
  const dangerousResetCmds = [
    "git reset --hard main",
    "git reset --hard master",
    "git reset --hard origin/main",
  ];

  for (const cmd of dangerousResetCmds) {
    it(`blocks ${cmd}`, () => {
      const result = evaluateRules(makeCtx("Bash", { command: cmd }));
      expect(result.decision).toBe("deny");
    });
  }
});

// ============================================================
// R12: direct push to protected branch warning
// ============================================================
describe("R12: direct push to protected branch warning", () => {
  it("git push origin main returns approve + systemMessage", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "git push origin main" })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeTruthy();
    expect(result.systemMessage).toContain("main");
  });

  it("git push upstream master returns approve + systemMessage", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "git push upstream master" })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeTruthy();
    expect(result.systemMessage).toContain("master");
  });
});

// ============================================================
// R13: important file change warning
// ============================================================
describe("R13: important file change warning", () => {
  const protectedPaths = [
    "package.json",
    "Dockerfile",
    "docker-compose.yml",
    ".github/workflows/ci.yml",
    "schema.prisma",
    "wrangler.toml",
    "index.html",
  ];

  for (const path of protectedPaths) {
    it(`Write to ${path} returns approve + systemMessage`, () => {
      const result = evaluateRules(
        makeCtx("Write", { file_path: path })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeTruthy();
      expect(result.systemMessage).toContain(path);
    });
  }

  it("normal source file change does not warn", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "src/index.ts" })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeUndefined();
  });
});

// ============================================================
// R07: Write/Edit block in Codex mode
// ============================================================
describe("R07: Write/Edit block in Codex mode", () => {
  it("blocks Write in Codex mode", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "/project/src/foo.ts" }, { codexMode: true })
    );
    expect(result.decision).toBe("deny");
  });

  it("blocks Edit in Codex mode", () => {
    const result = evaluateRules(
      makeCtx("Edit", { file_path: "/project/src/foo.ts" }, { codexMode: true })
    );
    expect(result.decision).toBe("deny");
  });

  it("does not block Write in normal mode", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "/project/src/foo.ts" }, { codexMode: false })
    );
    expect(result.decision).toBe("approve");
  });

  it("does not block Bash in Codex mode (R07 toolPattern is Write/Edit only)", () => {
    // R07's toolPattern is /^(?:Write|Edit|MultiEdit)$/ only
    // evaluateRules checks toolPattern so Bash does not match R07
    const result = evaluateRules(
      makeCtx("Bash", { command: "ls" }, { codexMode: true })
    );
    expect(result.decision).toBe("approve");
  });
});

// ============================================================
// R08: Breezing reviewer role guard
// ============================================================
describe("R08: Breezing reviewer role guard", () => {
  it("blocks Write for reviewer role", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "/project/src/foo.ts" }, { breezingRole: "reviewer" })
    );
    expect(result.decision).toBe("deny");
  });

  it("blocks Edit for reviewer role", () => {
    const result = evaluateRules(
      makeCtx("Edit", { file_path: "/project/src/foo.ts" }, { breezingRole: "reviewer" })
    );
    expect(result.decision).toBe("deny");
  });

  it("blocks git commit for reviewer role", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "git commit -m 'test'" }, { breezingRole: "reviewer" })
    );
    expect(result.decision).toBe("deny");
  });

  it("blocks git push for reviewer role", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "git push origin main" }, { breezingRole: "reviewer" })
    );
    expect(result.decision).toBe("deny");
  });

  it("does not block ls for reviewer role (read-only command)", () => {
    const rule = GUARD_RULES.find((r) => r.id === "R08:breezing-reviewer-no-write")!;
    const result = rule.evaluate(
      makeCtx("Bash", { command: "ls -la" }, { breezingRole: "reviewer" })
    );
    // Does not match prohibited patterns, returns null
    expect(result).toBeNull();
  });

  it("does not block for non-reviewer role", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "/project/src/foo.ts" }, { breezingRole: "implementer" })
    );
    expect(result.decision).toBe("approve");
  });

  it("does not block when no role is set", () => {
    const result = evaluateRules(
      makeCtx("Write", { file_path: "/project/src/foo.ts" }, { breezingRole: null })
    );
    expect(result.decision).toBe("approve");
  });
});

// ============================================================
// R09: sensitive file Read warning (approve + systemMessage)
// ============================================================
describe("R09: sensitive file Read warning", () => {
  const secretPaths = [
    ".env",
    "id_rsa",
    "private.pem",
    "server.key",
    "secrets/api.json",
  ];

  for (const path of secretPaths) {
    it(`Read of ${path} returns approve + systemMessage`, () => {
      const result = evaluateRules(
        makeCtx("Read", { file_path: path })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toBeDefined();
      expect(result.systemMessage).toContain(path);
    });
  }

  it("Read of normal source file has no warning", () => {
    const result = evaluateRules(
      makeCtx("Read", { file_path: "src/index.ts" })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeUndefined();
  });
});

// ============================================================
// evaluateRules: integration tests
// ============================================================
describe("evaluateRules: integration tests", () => {
  it("skips rules when tool_input.command is not a string", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: 12345 })
    );
    expect(result.decision).toBe("approve");
  });

  it("returns approve when no rules match", () => {
    const result = evaluateRules(
      makeCtx("Bash", { command: "echo hello" })
    );
    expect(result.decision).toBe("approve");
  });

  it("Codex MCP tools are not subject to Bash rules", () => {
    // mcp__codex__* is not in GUARD_RULES (out of scope for rules.ts)
    // Blocked separately in index.ts
    const result = evaluateRules(
      makeCtx("mcp__codex__exec", { input: "ls" })
    );
    expect(result.decision).toBe("approve");
  });

  it("R01 takes priority when both R01 and R05 match", () => {
    // sudo + rm -rf command → R01 matches first
    const result = evaluateRules(
      makeCtx("Bash", { command: "sudo rm -rf /" })
    );
    expect(result.decision).toBe("deny");
    // deny includes R01's description
    expect(result.reason).toContain("sudo");
  });
});
