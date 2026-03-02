/**
 * core/src/state/__tests__/store.test.ts
 * HarnessStore 単体テスト
 *
 * better-sqlite3 を使った実際の SQLite DB（メモリ内）で各メソッドを検証する。
 */

import { beforeEach, afterEach, describe, it, expect } from "vitest";
import Database from "better-sqlite3";
import { HarnessStore } from "../store.js";

// テスト用に HarnessStore のサブクラスを作り、メモリ DB を使う
// HarnessStore のコンストラクタはファイルパスを受け取るが、
// ":memory:" を渡すことで better-sqlite3 はメモリ内 DB を使用する
function createStore(): HarnessStore {
  return new HarnessStore(":memory:");
}

describe("HarnessStore", () => {
  let store: HarnessStore;

  beforeEach(() => {
    store = createStore();
  });

  afterEach(() => {
    store.close();
  });

  // ------------------------------------------------------------------
  // セッション管理
  // ------------------------------------------------------------------

  describe("upsertSession / getSession", () => {
    it("セッションを登録して取得できる", () => {
      store.upsertSession({
        session_id: "sess-001",
        mode: "work",
        project_root: "/tmp/project",
        started_at: "2026-01-01T00:00:00Z",
      });

      const session = store.getSession("sess-001");
      expect(session).not.toBeNull();
      expect(session?.session_id).toBe("sess-001");
      expect(session?.mode).toBe("work");
      expect(session?.project_root).toBe("/tmp/project");
    });

    it("存在しないセッションは null を返す", () => {
      const session = store.getSession("nonexistent");
      expect(session).toBeNull();
    });

    it("upsert で既存セッションを更新できる", () => {
      store.upsertSession({
        session_id: "sess-002",
        mode: "normal",
        project_root: "/tmp/project",
        started_at: "2026-01-01T00:00:00Z",
      });

      store.upsertSession({
        session_id: "sess-002",
        mode: "codex",
        project_root: "/tmp/project-v2",
        started_at: "2026-01-01T00:00:00Z",
      });

      const session = store.getSession("sess-002");
      expect(session?.mode).toBe("codex");
      expect(session?.project_root).toBe("/tmp/project-v2");
    });

    it("endSession でセッションを終了できる", () => {
      store.upsertSession({
        session_id: "sess-003",
        mode: "breezing",
        project_root: "/tmp",
        started_at: "2026-01-01T00:00:00Z",
      });

      store.endSession("sess-003");

      const session = store.getSession("sess-003");
      expect(session).not.toBeNull();
      // ended_at は getSession ではマッピングされないが、
      // セッション自体は取得できる
    });
  });

  // ------------------------------------------------------------------
  // シグナル管理
  // ------------------------------------------------------------------

  describe("sendSignal / receiveSignals", () => {
    it("シグナルを送信して受信できる", () => {
      store.upsertSession({
        session_id: "sender",
        mode: "normal",
        project_root: "/tmp",
        started_at: new Date().toISOString(),
      });
      store.upsertSession({
        session_id: "receiver",
        mode: "normal",
        project_root: "/tmp",
        started_at: new Date().toISOString(),
      });

      store.sendSignal({
        type: "task_completed",
        from_session_id: "sender",
        to_session_id: "receiver",
        payload: { task_id: "task-01", result: "success" },
      });

      const signals = store.receiveSignals("receiver");
      expect(signals).toHaveLength(1);
      expect(signals[0]?.type).toBe("task_completed");
      expect(signals[0]?.from_session_id).toBe("sender");
      expect(signals[0]?.payload).toEqual({ task_id: "task-01", result: "success" });
    });

    it("ブロードキャストシグナルはすべてのセッションが受信できる", () => {
      store.upsertSession({
        session_id: "broadcaster",
        mode: "normal",
        project_root: "/tmp",
        started_at: new Date().toISOString(),
      });

      store.sendSignal({
        type: "teammate_idle",
        from_session_id: "broadcaster",
        payload: {},
      });

      const signals = store.receiveSignals("any-session");
      expect(signals).toHaveLength(1);
      expect(signals[0]?.type).toBe("teammate_idle");
    });

    it("受信済みシグナルは再度受信されない", () => {
      store.upsertSession({
        session_id: "s1",
        mode: "normal",
        project_root: "/tmp",
        started_at: new Date().toISOString(),
      });

      store.sendSignal({
        type: "session_start",
        from_session_id: "s1",
        to_session_id: "s2",
        payload: {},
      });

      const first = store.receiveSignals("s2");
      expect(first).toHaveLength(1);

      const second = store.receiveSignals("s2");
      expect(second).toHaveLength(0);
    });

    it("自分が送ったシグナルは自分では受信しない", () => {
      store.sendSignal({
        type: "request_review",
        from_session_id: "s1",
        payload: {},
      });

      const signals = store.receiveSignals("s1");
      expect(signals).toHaveLength(0);
    });
  });

  // ------------------------------------------------------------------
  // タスク失敗管理
  // ------------------------------------------------------------------

  describe("recordFailure / getFailures", () => {
    it("タスク失敗を記録して取得できる", () => {
      const id = store.recordFailure(
        {
          task_id: "task-01",
          severity: "error",
          message: "TypeScript compilation failed",
          attempt: 1,
        },
        "sess-001"
      );

      expect(id).toBeGreaterThan(0);

      const failures = store.getFailures("task-01");
      expect(failures).toHaveLength(1);
      expect(failures[0]?.severity).toBe("error");
      expect(failures[0]?.message).toBe("TypeScript compilation failed");
      expect(failures[0]?.attempt).toBe(1);
    });

    it("detail フィールドはオプション（省略可能）", () => {
      store.recordFailure(
        {
          task_id: "task-02",
          severity: "warning",
          message: "Minor issue",
          attempt: 1,
        },
        "sess-001"
      );

      const failures = store.getFailures("task-02");
      expect(failures[0]?.detail).toBeUndefined();
    });

    it("detail フィールドがある場合は取得できる", () => {
      store.recordFailure(
        {
          task_id: "task-03",
          severity: "critical",
          message: "Fatal error",
          detail: "Stack trace here",
          attempt: 2,
        },
        "sess-001"
      );

      const failures = store.getFailures("task-03");
      expect(failures[0]?.detail).toBe("Stack trace here");
    });

    it("存在しないタスクは空配列を返す", () => {
      const failures = store.getFailures("nonexistent-task");
      expect(failures).toHaveLength(0);
    });
  });

  // ------------------------------------------------------------------
  // work_states 管理
  // ------------------------------------------------------------------

  describe("setWorkState / getWorkState / cleanExpiredWorkStates", () => {
    it("work state を設定して取得できる", () => {
      // FK 制約を満たすためにセッションを事前登録
      store.upsertSession({
        session_id: "sess-001",
        mode: "work",
        project_root: "/tmp",
        started_at: new Date().toISOString(),
      });
      store.setWorkState("sess-001", {
        codexMode: true,
        bypassRmRf: false,
        bypassGitPush: false,
      });

      const state = store.getWorkState("sess-001");
      expect(state).not.toBeNull();
      expect(state?.codexMode).toBe(true);
      expect(state?.bypassRmRf).toBe(false);
      expect(state?.bypassGitPush).toBe(false);
    });

    it("デフォルト値はすべて false", () => {
      // FK 制約を満たすためにセッションを事前登録
      store.upsertSession({
        session_id: "sess-002",
        mode: "normal",
        project_root: "/tmp",
        started_at: new Date().toISOString(),
      });
      store.setWorkState("sess-002");

      const state = store.getWorkState("sess-002");
      expect(state?.codexMode).toBe(false);
      expect(state?.bypassRmRf).toBe(false);
      expect(state?.bypassGitPush).toBe(false);
    });

    it("存在しない session の work state は null", () => {
      const state = store.getWorkState("nonexistent");
      expect(state).toBeNull();
    });

    it("cleanExpiredWorkStates で期限切れを削除できる", () => {
      // FK 制約を満たすためにセッションを事前登録
      store.upsertSession({
        session_id: "expired-sess",
        mode: "normal",
        project_root: "/tmp",
        started_at: new Date().toISOString(),
      });
      // DB に直接期限切れレコードを挿入
      const db = (store as unknown as { db: InstanceType<typeof Database> }).db;
      const expiredAt = Math.floor(Date.now() / 1000) - 1; // 1秒前 = 期限切れ
      db.prepare(
        `INSERT INTO work_states(session_id, codex_mode, bypass_rm_rf, bypass_git_push, expires_at)
         VALUES ('expired-sess', 0, 0, 0, ?)`
      ).run(expiredAt);

      const deleted = store.cleanExpiredWorkStates();
      expect(deleted).toBe(1);

      const state = store.getWorkState("expired-sess");
      expect(state).toBeNull();
    });
  });
});
