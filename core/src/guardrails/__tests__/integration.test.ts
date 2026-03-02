/**
 * core/src/guardrails/__tests__/integration.test.ts
 * Harness v3 ガードレール E2E 統合テスト
 *
 * 実際のフック呼び出しフロー（evaluatePreTool → evaluatePostTool）を通じて
 * 9つのガードルールが連携して正しく動作することを検証する。
 *
 * ユニットテストとの違い:
 *   - rules.test.ts: 個別ルール関数の単体テスト
 *   - integration.test.ts: PreToolUse → PostToolUse の実フロー全体テスト
 */

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { evaluatePreTool } from "../pre-tool.js";
import { evaluatePostTool } from "../post-tool.js";
import type { HookInput } from "../../types.js";

// ============================================================
// テストヘルパー
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
// PreToolUse → 判定結果の統合テスト
// ============================================================

describe("E2E: PreToolUse フロー", () => {
  // 各テストで環境変数をリセット
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
  // 正常系: 問題のない操作は approve される
  // ------------------------------------------------------------------

  describe("approve ケース", () => {
    it("通常の Bash コマンドは approve される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "npm test" })
      );
      expect(result.decision).toBe("approve");
    });

    it("通常のファイル書き込みは approve される", async () => {
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/src/index.ts" })
      );
      expect(result.decision).toBe("approve");
    });

    it("rm -r フラグなしの削除は approve される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "rm /tmp/test.log" })
      );
      expect(result.decision).toBe("approve");
    });

    it("git push（force なし）は approve される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push origin main" })
      );
      expect(result.decision).toBe("approve");
    });

    it("通常の Read は approve される", async () => {
      const result = await evaluatePreTool(
        buildInput("Read", { file_path: "/test/project/src/app.ts" })
      );
      expect(result.decision).toBe("approve");
    });
  });

  // ------------------------------------------------------------------
  // deny ケース: 危険操作は阻止される
  // ------------------------------------------------------------------

  describe("deny ケース", () => {
    it("sudo は deny される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "sudo apt-get update" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("sudo");
    });

    it("git push --force は deny される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push --force origin main" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("force");
    });

    it("git push -f も deny される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push -f origin main" })
      );
      expect(result.decision).toBe("deny");
    });

    it(".env ファイルへの Write は deny される", async () => {
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/.env" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain(".env");
    });

    it(".git/ ディレクトリへの Edit は deny される", async () => {
      const result = await evaluatePreTool(
        buildInput("Edit", { file_path: "/test/project/.git/config" })
      );
      expect(result.decision).toBe("deny");
    });

    it("Bash での .env への書き込みは deny される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "echo 'SECRET=123' > .env" })
      );
      expect(result.decision).toBe("deny");
    });

    it("Bash での .env バリアントへの書き込みも deny される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "echo 'KEY=val' >> .env.local" })
      );
      expect(result.decision).toBe("deny");
    });

    it("秘密鍵ファイルへの Write は deny される", async () => {
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/root/.ssh/id_rsa" })
      );
      expect(result.decision).toBe("deny");
    });

    it("codex モード時の Write は deny される", async () => {
      process.env["HARNESS_CODEX_MODE"] = "1";
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/src/app.ts" })
      );
      expect(result.decision).toBe("deny");
      expect(result.reason).toContain("Codex");
    });

    it("breezing reviewer は git commit を deny される", async () => {
      process.env["HARNESS_BREEZING_ROLE"] = "reviewer";
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git commit -m 'feat: add feature'" })
      );
      expect(result.decision).toBe("deny");
    });

    it("breezing reviewer は Write を deny される", async () => {
      process.env["HARNESS_BREEZING_ROLE"] = "reviewer";
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/test/project/src/app.ts" })
      );
      expect(result.decision).toBe("deny");
    });
  });

  // ------------------------------------------------------------------
  // ask ケース: 確認が必要な操作
  // ------------------------------------------------------------------

  describe("ask ケース（work モード以外）", () => {
    it("rm -rf は ask される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "rm -rf /tmp/test-dir" })
      );
      expect(result.decision).toBe("ask");
      expect(result.reason).toContain("rm");
    });

    it("rm -fr も ask される", async () => {
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "rm -fr dist/" })
      );
      expect(result.decision).toBe("ask");
    });

    it("プロジェクト外の絶対パスへの Write は ask される", async () => {
      process.env["HARNESS_PROJECT_ROOT"] = "/test/project";
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/etc/hosts" })
      );
      expect(result.decision).toBe("ask");
    });
  });

  // ------------------------------------------------------------------
  // work モード: バイパス可能な操作
  // ------------------------------------------------------------------

  describe("work モード バイパスケース", () => {
    it("work モード時は rm -rf が ask されない", async () => {
      process.env["HARNESS_WORK_MODE"] = "1";
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "rm -rf dist/" })
      );
      // work モードでは ask をスキップして approve になる
      expect(result.decision).toBe("approve");
    });

    it("work モード時もプロジェクト外への書き込みは ask されない", async () => {
      process.env["HARNESS_WORK_MODE"] = "1";
      process.env["HARNESS_PROJECT_ROOT"] = "/test/project";
      const result = await evaluatePreTool(
        buildInput("Write", { file_path: "/tmp/output.txt" })
      );
      expect(result.decision).toBe("approve");
    });

    it("work モードでも sudo は deny される（例外なし）", async () => {
      process.env["HARNESS_WORK_MODE"] = "1";
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "sudo make install" })
      );
      expect(result.decision).toBe("deny");
    });

    it("work モードでも git push --force は deny される（例外なし）", async () => {
      process.env["HARNESS_WORK_MODE"] = "1";
      const result = await evaluatePreTool(
        buildInput("Bash", { command: "git push --force origin main" })
      );
      expect(result.decision).toBe("deny");
    });
  });

  // ------------------------------------------------------------------
  // approve + systemMessage ケース: 警告付き承認
  // ------------------------------------------------------------------

  describe("approve + 警告 ケース", () => {
    it(".env ファイルの Read は approve されるが警告が出る", async () => {
      const result = await evaluatePreTool(
        buildInput("Read", { file_path: "/test/project/.env" })
      );
      expect(result.decision).toBe("approve");
      expect(result.systemMessage).toContain("警告");
      expect(result.systemMessage).toContain(".env");
    });
  });
});

// ============================================================
// PostToolUse フロー
// ============================================================

describe("E2E: PostToolUse フロー", () => {
  beforeEach(() => {
    delete process.env["HARNESS_WORK_MODE"];
  });

  it("通常の Write 結果は approve される", async () => {
    const result = await evaluatePostTool(
      buildInput("Write", {
        file_path: "/test/project/src/app.ts",
        content: "export const app = {};",
      })
    );
    expect(result.decision).toBe("approve");
  });

  it("テスト改ざんを検出すると approve + 警告が返る", async () => {
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

  it("ESLint disable コメントを検出すると警告が出る", async () => {
    const result = await evaluatePostTool(
      buildInput("Write", {
        file_path: "/test/project/.eslintrc.js",
        content: "/* eslint-disable */\nmodule.exports = {};",
      })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeTruthy();
  });

  it("CI の continue-on-error 追加を検出すると警告が出る", async () => {
    const result = await evaluatePostTool(
      buildInput("Write", {
        file_path: "/test/project/.github/workflows/ci.yml",
        content: "continue-on-error: true",
      })
    );
    expect(result.decision).toBe("approve");
    expect(result.systemMessage).toBeTruthy();
  });

  it("通常の Bash 実行結果は approve される", async () => {
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
// プリ → ポスト の連続フロー
// ============================================================

describe("E2E: PreToolUse → PostToolUse 連続フロー", () => {
  it("正常な Write 操作は両フックで approve される", async () => {
    const input = buildInput("Write", {
      file_path: "/test/project/src/feature.ts",
      content: "export const feature = () => 'hello';",
    });

    const preResult = await evaluatePreTool(input);
    expect(preResult.decision).toBe("approve");

    const postResult = await evaluatePostTool(input);
    expect(postResult.decision).toBe("approve");
  });

  it("sudo は PreToolUse で阻止される（PostToolUse は不要）", async () => {
    const input = buildInput("Bash", { command: "sudo apt-get install curl" });

    const preResult = await evaluatePreTool(input);
    expect(preResult.decision).toBe("deny");
    // deny された場合、PostToolUse は実行されない（ここではシミュレーションのみ）
    expect(preResult.reason).toContain("sudo");
  });
});
