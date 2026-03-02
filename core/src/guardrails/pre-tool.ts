/**
 * core/src/guardrails/pre-tool.ts
 * PreToolUse フック評価関数
 *
 * HookInput を受け取り、rules.ts の宣言的ガードルールテーブルを評価して
 * approve / deny / ask の HookResult を返す。
 */

import { type HookInput, type HookResult, type RuleContext } from "../types.js";
import { evaluateRules } from "./rules.js";

/**
 * 実行環境から RuleContext を組み立てる。
 * work-active.json / session-state の読み取りは Phase 17.2 で SQLite に移行予定。
 * 現時点では環境変数・HookInput の cwd / plugin_root からコンテキストを取得する。
 */
/** 環境変数が truthy 値（"1", "true", "yes"）かどうか判定 */
function isTruthy(value: string | undefined): boolean {
  return value === "1" || value === "true" || value === "yes";
}

function buildContext(input: HookInput): RuleContext {
  // cwd がプロジェクトルート。plugin_root はプラグイン自身のパスなので除外
  const projectRoot =
    input.cwd ??
    process.env["HARNESS_PROJECT_ROOT"] ??
    process.env["PROJECT_ROOT"] ??
    process.cwd();

  // work モード: 環境変数または work-active.json を参照（簡易実装）
  const workMode =
    isTruthy(process.env["HARNESS_WORK_MODE"]) ||
    isTruthy(process.env["ULTRAWORK_MODE"]);

  // codex モード: 環境変数から取得
  const codexMode = isTruthy(process.env["HARNESS_CODEX_MODE"]);

  // breezing ロール: 環境変数から取得
  const breezingRole = process.env["HARNESS_BREEZING_ROLE"] ?? null;

  return {
    input,
    projectRoot,
    workMode,
    codexMode,
    breezingRole,
  };
}

/**
 * PreToolUse フックのエントリポイント。
 * HookInput を受け取り、ガードルールを評価して HookResult を返す。
 */
export function evaluatePreTool(input: HookInput): HookResult {
  const ctx = buildContext(input);
  return evaluateRules(ctx);
}
