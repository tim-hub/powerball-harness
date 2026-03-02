/**
 * core/src/guardrails/permission.ts
 * PermissionRequest フック評価関数
 *
 * permission-request.sh の全ロジックを TypeScript に移植。
 * 安全なコマンド（read-only git、テストコマンド等）を自動承認する。
 *
 * 参照元: scripts/permission-request.sh
 */

import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { type HookInput, type HookResult } from "../types.js";

// ============================================================
// PermissionRequest 固有の出力形式
// ============================================================

/** PermissionRequest フックの決定レスポンス */
interface PermissionResponse {
  hookSpecificOutput: {
    hookEventName: "PermissionRequest";
    decision: {
      behavior: "allow" | "deny";
    };
  };
}

function makeAllow(): PermissionResponse {
  return {
    hookSpecificOutput: {
      hookEventName: "PermissionRequest",
      decision: { behavior: "allow" },
    },
  };
}

// ============================================================
// パッケージマネージャー自動承認 許可リスト
// ============================================================

/**
 * .claude/config/allowed-pkg-managers.json が存在し、allowed: true であれば
 * npm/pnpm/yarn の test/build/lint 等を自動承認する。
 */
function isPkgManagerAllowed(cwd: string): boolean {
  const allowlistPath = join(cwd, ".claude", "config", "allowed-pkg-managers.json");
  if (!existsSync(allowlistPath)) return false;

  try {
    const raw = readFileSync(allowlistPath, "utf-8");
    const data = JSON.parse(raw) as unknown;
    if (typeof data === "object" && data !== null && "allowed" in data) {
      return (data as Record<string, unknown>)["allowed"] === true;
    }
  } catch {
    // JSON パースエラーは不許可として扱う
  }

  return false;
}

// ============================================================
// 安全なコマンド判定
// ============================================================

/**
 * コマンド文字列が自動承認可能かを判定する。
 *
 * security hardening:
 * - パイプ、リダイレクト、変数展開、コマンド置換を含む場合は不承認（保守的）
 * - シンプルなコマンドのみ自動承認
 */
function isSafeCommand(command: string, cwd: string): boolean {
  // 複数行コマンドは不承認
  if (command.includes("\n") || command.includes("\r")) return false;

  // シェル特殊文字（パイプ、リダイレクト、変数展開、コマンド置換）を含む場合は不承認
  if (/[;&|<>`$]/.test(command)) return false;

  // read-only git コマンドは常に安全
  if (/^git\s+(status|diff|log|branch|rev-parse|show|ls-files)(\s|$)/i.test(command)) {
    return true;
  }

  // JS/TS テスト・検証コマンドはパッケージマネージャー許可リストを確認
  if (
    /^(npm|pnpm|yarn)\s+(test|run\s+(test|lint|typecheck|build|validate)|lint|typecheck|build)(\s|$)/i.test(
      command
    )
  ) {
    return isPkgManagerAllowed(cwd);
  }

  // Python テスト（package.json リスクなし）
  if (/^(pytest|python\s+-m\s+pytest)(\s|$)/i.test(command)) return true;

  // Go / Rust テスト
  if (/^(go\s+test|cargo\s+test)(\s|$)/i.test(command)) return true;

  return false;
}

// ============================================================
// evaluatePermission: メインエクスポート
// ============================================================

/**
 * PermissionRequest フックの評価関数。
 *
 * Edit/Write は bypassPermissions 相当で自動承認。
 * Bash は安全なコマンドパターンのみ自動承認。
 * その他は何も返さず（デフォルト動作 = ユーザーに確認）。
 */
export function evaluatePermission(input: HookInput): HookResult {
  const toolName = input.tool_name;
  const cwd = input.cwd ?? process.cwd();

  // Edit / Write は自動承認（bypassPermissions モード補完）
  if (toolName === "Edit" || toolName === "Write" || toolName === "MultiEdit") {
    return _permissionResponseToHookResult(makeAllow());
  }

  // Bash 以外はデフォルト動作（スルー）
  if (toolName !== "Bash") {
    return { decision: "approve" };
  }

  // Bash: コマンドを取得して安全性チェック
  const command = input.tool_input["command"];
  if (typeof command !== "string" || command.trim() === "") {
    return { decision: "approve" };
  }

  if (isSafeCommand(command, cwd)) {
    return _permissionResponseToHookResult(makeAllow());
  }

  // 安全でないコマンドはデフォルト動作（ユーザーに確認を委ねる）
  return { decision: "approve" };
}

/**
 * PermissionResponse を HookResult に変換する。
 *
 * PermissionRequest フックは通常の HookResult とは異なる出力形式だが、
 * 内部の型システムでは HookResult として扱い、
 * index.ts の route() で stdout 出力時に formatPermissionOutput() を使って
 * 正しい形式に変換する（Phase 17.1.7 で hooks.json を差し替え後に対応予定）。
 */
function _permissionResponseToHookResult(response: PermissionResponse): HookResult {
  return {
    decision: "approve",
    systemMessage: JSON.stringify(response),
  };
}

/**
 * PermissionRequest フック用の stdout 出力を生成する。
 * index.ts の route() から "permission" フックタイプ時に呼び出す。
 */
export function formatPermissionOutput(result: HookResult): string {
  // systemMessage に PermissionResponse の JSON が入っている場合はそちらを優先
  if (result.systemMessage !== undefined) {
    try {
      const parsed = JSON.parse(result.systemMessage) as unknown;
      if (
        typeof parsed === "object" &&
        parsed !== null &&
        "hookSpecificOutput" in parsed
      ) {
        return JSON.stringify(parsed);
      }
    } catch {
      // パース失敗時は通常の HookResult として出力
    }
  }

  return JSON.stringify(result);
}
