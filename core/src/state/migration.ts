/**
 * core/src/state/migration.ts
 * Harness v2 JSON / JSONL → v3 SQLite 移行スクリプト
 *
 * v2 の状態ファイルを v3 SQLite DB に取り込む。
 * 移行対象:
 *   .claude/state/session.json      → sessions テーブル
 *   .claude/state/session.events.jsonl → signals テーブル（task_completed 等）
 *   .claude/work-active.json         → work_states テーブル
 *
 * 冪等設計: 既に移行済みの場合は再実行しても安全。
 */

import { readFileSync, existsSync, renameSync } from "node:fs";
import { resolve } from "node:path";
import type { SignalType } from "../types.js";
import { HarnessStore } from "./store.js";

// ============================================================
// 型定義（v2 JSON 構造）
// ============================================================

interface V2Session {
  session_id?: string;
  id?: string; // 旧フィールド名
  mode?: string;
  project_root?: string;
  started_at?: string | number;
  ended_at?: string | number | null;
  context?: Record<string, unknown>;
}

interface V2Event {
  type?: string;
  event?: string; // 旧フィールド名
  session_id?: string;
  from_session_id?: string;
  to_session_id?: string | null;
  payload?: Record<string, unknown>;
  data?: Record<string, unknown>; // 旧フィールド名
  timestamp?: string | number;
  sent_at?: string | number; // 旧フィールド名
}

interface V2WorkActive {
  session_id?: string;
  codex_mode?: boolean;
  bypass_rm_rf?: boolean;
  bypass_git_push?: boolean;
  mode?: string;
}

// ============================================================
// ヘルパー関数
// ============================================================

/** ISO 日付文字列または Unix タイムスタンプを ISO 文字列に正規化 */
function toIsoString(value: string | number | null | undefined): string {
  if (value === null || value === undefined) {
    return new Date().toISOString();
  }
  if (typeof value === "number") {
    // Unix タイムスタンプ秒またはミリ秒を判定
    const ms = value > 1e10 ? value : value * 1000;
    return new Date(ms).toISOString();
  }
  // 既に ISO 文字列の場合はそのまま返す
  return value;
}

/** v2 モード文字列を v3 モードに正規化 */
function normalizeMode(mode: string | undefined): "normal" | "work" | "codex" | "breezing" {
  switch (mode) {
    case "work":
    case "codex":
    case "breezing":
      return mode;
    default:
      return "normal";
  }
}

/** SignalType として有効な文字列かチェック */
function normalizeSignalType(type: string | undefined): SignalType {
  // 有効な SignalType 一覧（types.ts の SignalType と同期）
  const valid: SignalType[] = [
    "task_completed", "task_failed", "teammate_idle",
    "session_start", "session_end", "stop_failure", "request_review",
  ];
  if (type && (valid as string[]).includes(type)) return type as SignalType;
  return "task_completed"; // 不明な型はフォールバック
}

// ============================================================
// JSON ファイル読み込みユーティリティ
// ============================================================

/** JSON ファイルを安全に読み込む。存在しない場合は null を返す */
function readJsonFile<T>(filePath: string): T | null {
  if (!existsSync(filePath)) return null;
  try {
    const content = readFileSync(filePath, "utf8");
    return JSON.parse(content) as T;
  } catch {
    return null;
  }
}

/** JSONL ファイルを安全に読み込む（1行1JSON）。存在しない場合は [] を返す */
function readJsonlFile<T>(filePath: string): T[] {
  if (!existsSync(filePath)) return [];
  try {
    const content = readFileSync(filePath, "utf8");
    return content
      .split("\n")
      .filter((line) => line.trim().length > 0)
      .map((line) => JSON.parse(line) as T);
  } catch {
    return [];
  }
}

// ============================================================
// 移行処理
// ============================================================

export interface MigrationResult {
  sessions: number;
  signals: number;
  workStates: number;
  skipped: boolean;
  errors: string[];
}

/**
 * v2 JSON/JSONL 状態ファイルを v3 SQLite DB に移行する。
 *
 * @param projectRoot - プロジェクトルートのパス（デフォルト: process.cwd()）
 * @param dbPath - SQLite DB のパス（デフォルト: <projectRoot>/.harness/state.db）
 * @returns 移行結果
 */
export function migrate(
  projectRoot: string = process.cwd(),
  dbPath?: string
): MigrationResult {
  const stateDir = resolve(projectRoot, ".claude", "state");
  const resolvedDbPath = dbPath ?? resolve(projectRoot, ".harness", "state.db");

  const result: MigrationResult = {
    sessions: 0,
    signals: 0,
    workStates: 0,
    skipped: false,
    errors: [],
  };

  const store = new HarnessStore(resolvedDbPath);

  try {
    // 移行済みチェック: schema_meta に migration_done が存在すれば スキップ
    const migrationDone = store.getMeta("migration_v1_done");
    if (migrationDone === "1") {
      result.skipped = true;
      return result;
    }

    // ------------------------------------------------
    // 1. session.json → sessions テーブル
    // ------------------------------------------------
    const sessionFile = resolve(stateDir, "session.json");
    const v2Session = readJsonFile<V2Session>(sessionFile);

    if (v2Session !== null) {
      const sessionId = v2Session.session_id ?? v2Session.id ?? "migrated-session";
      try {
        store.upsertSession({
          session_id: sessionId,
          mode: normalizeMode(v2Session.mode),
          project_root: v2Session.project_root ?? projectRoot,
          started_at: toIsoString(v2Session.started_at),
        });
        if (v2Session.ended_at !== null && v2Session.ended_at !== undefined) {
          store.endSession(sessionId);
        }
        result.sessions++;
      } catch (err) {
        result.errors.push(`session migration failed: ${err}`);
      }
    }

    // ------------------------------------------------
    // 2. session.events.jsonl → signals テーブル
    // ------------------------------------------------
    const eventsFile = resolve(stateDir, "session.events.jsonl");
    const v2Events = readJsonlFile<V2Event>(eventsFile);

    for (const event of v2Events) {
      const type = normalizeSignalType(event.type ?? event.event);
      const fromSessionId = event.from_session_id ?? event.session_id ?? "unknown";
      const payload = event.payload ?? event.data ?? {};

      try {
        const signal: Parameters<HarnessStore["sendSignal"]>[0] = {
          type,
          from_session_id: fromSessionId,
          payload,
        };
        if (event.to_session_id) {
          signal.to_session_id = event.to_session_id;
        }
        store.sendSignal(signal);
        result.signals++;
      } catch (err) {
        result.errors.push(`signal migration failed (type=${type}): ${err}`);
      }
    }

    // ------------------------------------------------
    // 3. work-active.json → work_states テーブル
    // ------------------------------------------------
    const workActiveFile = resolve(projectRoot, ".claude", "work-active.json");
    const v2WorkActive = readJsonFile<V2WorkActive>(workActiveFile);

    if (v2WorkActive !== null) {
      const sessionId = v2WorkActive.session_id ?? "migrated-work-session";
      try {
        // FK 制約を満たすため sessions に仮登録
        store.upsertSession({
          session_id: sessionId,
          mode: normalizeMode(v2WorkActive.mode ?? "work"),
          project_root: projectRoot,
          started_at: new Date().toISOString(),
        });
        store.setWorkState(sessionId, {
          codexMode: v2WorkActive.codex_mode ?? false,
          bypassRmRf: v2WorkActive.bypass_rm_rf ?? false,
          bypassGitPush: v2WorkActive.bypass_git_push ?? false,
        });
        result.workStates++;
      } catch (err) {
        result.errors.push(`work_state migration failed: ${err}`);
      }
    }

    // ------------------------------------------------
    // 4. 移行完了マークを記録
    // ------------------------------------------------
    store.setMeta("migration_v1_done", "1");

    // ------------------------------------------------
    // 5. 元ファイルをバックアップ（削除はしない）
    // ------------------------------------------------
    if (v2Session !== null && existsSync(sessionFile)) {
      try {
        renameSync(sessionFile, `${sessionFile}.v2.bak`);
      } catch {
        // バックアップ失敗は無視（移行自体は完了済み）
      }
    }

  } finally {
    store.close();
  }

  return result;
}

// ============================================================
// CLI エントリポイント（node で直接実行された場合）
// ============================================================

// ESM では import.meta.url で「直接実行」を判定できる
// dist/ にコンパイルされた後は `node dist/state/migration.js` で呼ぶ
const isMain = process.argv[1]?.endsWith("migration.js");

if (isMain) {
  const projectRoot = process.argv[2] ?? process.cwd();
  const dbPath = process.argv[3];

  const result = migrate(projectRoot, dbPath);

  if (result.skipped) {
    console.log("Migration already completed. Skipped.");
    process.exit(0);
  }

  if (result.errors.length > 0) {
    console.error("Migration completed with errors:");
    for (const err of result.errors) {
      console.error(`  - ${err}`);
    }
  }

  console.log(
    `Migration done: ${result.sessions} sessions, ${result.signals} signals, ${result.workStates} work_states`
  );
  process.exit(result.errors.length > 0 ? 1 : 0);
}
