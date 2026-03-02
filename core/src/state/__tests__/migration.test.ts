/**
 * core/src/state/__tests__/migration.test.ts
 * migration.ts の単体テスト
 *
 * 実際のファイルシステム操作を含むため、tmp ディレクトリを使用する。
 */

import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { mkdirSync, writeFileSync, existsSync, rmSync } from "node:fs";
import { resolve, join } from "node:path";
import { tmpdir } from "node:os";
import { migrate } from "../migration.js";
import { HarnessStore } from "../store.js";

// ============================================================
// テストユーティリティ
// ============================================================

/** 一時ディレクトリを作成して返す */
function createTmpProject(): string {
  const dir = join(tmpdir(), `harness-migration-test-${Date.now()}-${Math.random().toString(36).slice(2)}`);
  mkdirSync(dir, { recursive: true });
  mkdirSync(join(dir, ".claude", "state"), { recursive: true });
  mkdirSync(join(dir, ".harness"), { recursive: true });
  return dir;
}

/** テスト後に一時ディレクトリを削除する */
function cleanupTmpProject(dir: string): void {
  if (existsSync(dir)) {
    rmSync(dir, { recursive: true, force: true });
  }
}

// ============================================================
// テスト
// ============================================================

describe("migrate()", () => {
  let projectRoot: string;
  let dbPath: string;

  beforeEach(() => {
    projectRoot = createTmpProject();
    dbPath = join(projectRoot, ".harness", "state.db");
  });

  afterEach(() => {
    cleanupTmpProject(projectRoot);
  });

  // ------------------------------------------------------------------
  // 移行済みチェック
  // ------------------------------------------------------------------

  describe("移行済みスキップ", () => {
    it("移行済みの場合は skipped: true を返す", () => {
      // 1回目の移行
      const first = migrate(projectRoot, dbPath);
      expect(first.skipped).toBe(false);

      // 2回目は スキップ
      const second = migrate(projectRoot, dbPath);
      expect(second.skipped).toBe(true);
      expect(second.sessions).toBe(0);
      expect(second.signals).toBe(0);
    });
  });

  // ------------------------------------------------------------------
  // 状態ファイルなしの移行（空移行）
  // ------------------------------------------------------------------

  describe("空移行", () => {
    it("移行対象ファイルがない場合は 0件で完了する", () => {
      const result = migrate(projectRoot, dbPath);
      expect(result.skipped).toBe(false);
      expect(result.sessions).toBe(0);
      expect(result.signals).toBe(0);
      expect(result.workStates).toBe(0);
      expect(result.errors).toHaveLength(0);
    });
  });

  // ------------------------------------------------------------------
  // session.json の移行
  // ------------------------------------------------------------------

  describe("session.json 移行", () => {
    it("セッションを移行できる", () => {
      const sessionFile = resolve(projectRoot, ".claude", "state", "session.json");
      writeFileSync(sessionFile, JSON.stringify({
        session_id: "sess-migrate-01",
        mode: "work",
        project_root: projectRoot,
        started_at: "2026-01-01T00:00:00Z",
      }));

      const result = migrate(projectRoot, dbPath);
      expect(result.sessions).toBe(1);
      expect(result.errors).toHaveLength(0);

      // SQLite に保存されているか確認
      const store = new HarnessStore(dbPath);
      try {
        const session = store.getSession("sess-migrate-01");
        expect(session).not.toBeNull();
        expect(session?.session_id).toBe("sess-migrate-01");
        expect(session?.mode).toBe("work");
      } finally {
        store.close();
      }
    });

    it("Unix タイムスタンプ形式の started_at も移行できる", () => {
      const sessionFile = resolve(projectRoot, ".claude", "state", "session.json");
      writeFileSync(sessionFile, JSON.stringify({
        session_id: "sess-unix-ts",
        mode: "normal",
        project_root: projectRoot,
        started_at: 1704067200, // 2024-01-01T00:00:00Z
      }));

      const result = migrate(projectRoot, dbPath);
      expect(result.sessions).toBe(1);
      expect(result.errors).toHaveLength(0);

      const store = new HarnessStore(dbPath);
      try {
        const session = store.getSession("sess-unix-ts");
        expect(session).not.toBeNull();
      } finally {
        store.close();
      }
    });

    it("session_id が未設定でも移行できる（デフォルト ID を使用）", () => {
      const sessionFile = resolve(projectRoot, ".claude", "state", "session.json");
      writeFileSync(sessionFile, JSON.stringify({
        mode: "breezing",
        project_root: projectRoot,
        started_at: "2026-01-01T00:00:00Z",
      }));

      const result = migrate(projectRoot, dbPath);
      expect(result.sessions).toBe(1);
      expect(result.errors).toHaveLength(0);
    });

    it("無効な JSON の session.json は sessions: 0 で続行する", () => {
      const sessionFile = resolve(projectRoot, ".claude", "state", "session.json");
      writeFileSync(sessionFile, "{ invalid json }");

      const result = migrate(projectRoot, dbPath);
      expect(result.sessions).toBe(0);
      // エラーはないが session も 0（JSON パース失敗で null 扱い）
    });

    it("移行後に session.json が .v2.bak にリネームされる", () => {
      const sessionFile = resolve(projectRoot, ".claude", "state", "session.json");
      writeFileSync(sessionFile, JSON.stringify({
        session_id: "sess-backup-test",
        mode: "normal",
        project_root: projectRoot,
        started_at: "2026-01-01T00:00:00Z",
      }));

      migrate(projectRoot, dbPath);

      expect(existsSync(sessionFile)).toBe(false);
      expect(existsSync(`${sessionFile}.v2.bak`)).toBe(true);
    });
  });

  // ------------------------------------------------------------------
  // session.events.jsonl の移行
  // ------------------------------------------------------------------

  describe("session.events.jsonl 移行", () => {
    it("シグナルイベントを移行できる", () => {
      const eventsFile = resolve(projectRoot, ".claude", "state", "session.events.jsonl");
      const events = [
        { type: "task_completed", from_session_id: "sess-01", payload: { task: "impl" } },
        { type: "teammate_idle", from_session_id: "sess-02", payload: {} },
        { type: "session_start", from_session_id: "sess-03", to_session_id: "sess-04", payload: {} },
      ];
      writeFileSync(eventsFile, events.map(e => JSON.stringify(e)).join("\n"));

      const result = migrate(projectRoot, dbPath);
      expect(result.signals).toBe(3);
      expect(result.errors).toHaveLength(0);
    });

    it("空の JSONL ファイルは 0件で完了する", () => {
      const eventsFile = resolve(projectRoot, ".claude", "state", "session.events.jsonl");
      writeFileSync(eventsFile, "");

      const result = migrate(projectRoot, dbPath);
      expect(result.signals).toBe(0);
    });

    it("不明なイベントタイプはフォールバックシグナルに変換される", () => {
      const eventsFile = resolve(projectRoot, ".claude", "state", "session.events.jsonl");
      writeFileSync(eventsFile, JSON.stringify({
        type: "unknown_custom_event",
        from_session_id: "sess-01",
        payload: {},
      }));

      const result = migrate(projectRoot, dbPath);
      expect(result.signals).toBe(1);
      expect(result.errors).toHaveLength(0);
    });
  });

  // ------------------------------------------------------------------
  // work-active.json の移行
  // ------------------------------------------------------------------

  describe("work-active.json 移行", () => {
    it("work_state を移行できる", () => {
      const workActiveFile = resolve(projectRoot, ".claude", "work-active.json");
      writeFileSync(workActiveFile, JSON.stringify({
        session_id: "sess-work-01",
        mode: "work",
        codex_mode: true,
        bypass_rm_rf: false,
        bypass_git_push: false,
      }));

      const result = migrate(projectRoot, dbPath);
      expect(result.workStates).toBe(1);
      expect(result.errors).toHaveLength(0);

      const store = new HarnessStore(dbPath);
      try {
        const state = store.getWorkState("sess-work-01");
        expect(state).not.toBeNull();
        expect(state?.codexMode).toBe(true);
        expect(state?.bypassRmRf).toBe(false);
      } finally {
        store.close();
      }
    });
  });

  // ------------------------------------------------------------------
  // 複合移行（全ファイルが揃っている場合）
  // ------------------------------------------------------------------

  describe("複合移行", () => {
    it("session + events + work-active をすべて移行できる", () => {
      // session.json
      writeFileSync(
        resolve(projectRoot, ".claude", "state", "session.json"),
        JSON.stringify({
          session_id: "sess-full",
          mode: "codex",
          project_root: projectRoot,
          started_at: "2026-01-01T00:00:00Z",
        })
      );

      // events.jsonl
      writeFileSync(
        resolve(projectRoot, ".claude", "state", "session.events.jsonl"),
        [
          JSON.stringify({ type: "task_completed", from_session_id: "sess-full", payload: {} }),
          JSON.stringify({ type: "request_review", from_session_id: "sess-full", payload: {} }),
        ].join("\n")
      );

      // work-active.json
      writeFileSync(
        resolve(projectRoot, ".claude", "work-active.json"),
        JSON.stringify({
          session_id: "sess-full",
          mode: "codex",
          codex_mode: true,
          bypass_rm_rf: false,
          bypass_git_push: false,
        })
      );

      const result = migrate(projectRoot, dbPath);
      expect(result.sessions).toBe(1);
      expect(result.signals).toBe(2);
      expect(result.workStates).toBe(1);
      expect(result.errors).toHaveLength(0);
      expect(result.skipped).toBe(false);
    });
  });
});
