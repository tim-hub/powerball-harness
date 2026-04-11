/**
 * core/src/state/__tests__/store.test.ts
 * HarnessStore unit tests
 *
 * Tests each method using an actual SQLite DB (in-memory) via better-sqlite3.
 */

import { createRequire } from "node:module";
import { beforeEach, afterEach, describe, it, expect } from "vitest";
import { HarnessStore } from "../store.js";

const require = createRequire(import.meta.url);
const Database = require("better-sqlite3") as typeof import("better-sqlite3").default;

// Create a HarnessStore subclass for testing that uses an in-memory DB.
// HarnessStore's constructor takes a file path, but passing ":memory:"
// makes better-sqlite3 use an in-memory database.
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
  // Session management
  // ------------------------------------------------------------------

  describe("upsertSession / getSession", () => {
    it("can register and retrieve a session", () => {
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

    it("returns null for non-existent session", () => {
      const session = store.getSession("nonexistent");
      expect(session).toBeNull();
    });

    it("can update an existing session via upsert", () => {
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

    it("can end a session via endSession", () => {
      store.upsertSession({
        session_id: "sess-003",
        mode: "breezing",
        project_root: "/tmp",
        started_at: "2026-01-01T00:00:00Z",
      });

      store.endSession("sess-003");

      const session = store.getSession("sess-003");
      expect(session).not.toBeNull();
      // ended_at is not mapped by getSession, but
      // the session itself is still retrievable
    });
  });

  // ------------------------------------------------------------------
  // Signal management
  // ------------------------------------------------------------------

  describe("sendSignal / receiveSignals", () => {
    it("can send and receive signals", () => {
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

    it("broadcast signals can be received by all sessions", () => {
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

    it("already-received signals are not received again", () => {
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

    it("signals sent by self are not received by self", () => {
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
  // Task failure management
  // ------------------------------------------------------------------

  describe("recordFailure / getFailures", () => {
    it("can record and retrieve task failures", () => {
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

    it("detail field is optional", () => {
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

    it("detail field is retrievable when present", () => {
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

    it("returns empty array for non-existent task", () => {
      const failures = store.getFailures("nonexistent-task");
      expect(failures).toHaveLength(0);
    });
  });

  // ------------------------------------------------------------------
  // work_states management
  // ------------------------------------------------------------------

  describe("setWorkState / getWorkState / cleanExpiredWorkStates", () => {
    it("can set and retrieve work state", () => {
      // Pre-register session to satisfy FK constraint
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

    it("defaults are all false", () => {
      // Pre-register session to satisfy FK constraint
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

    it("returns null for non-existent session work state", () => {
      const state = store.getWorkState("nonexistent");
      expect(state).toBeNull();
    });

    it("can delete expired records via cleanExpiredWorkStates", () => {
      // Pre-register session to satisfy FK constraint
      store.upsertSession({
        session_id: "expired-sess",
        mode: "normal",
        project_root: "/tmp",
        started_at: new Date().toISOString(),
      });
      // Insert expired record directly into DB
      const db = (store as unknown as { db: InstanceType<typeof Database> }).db;
      const expiredAt = Math.floor(Date.now() / 1000) - 1; // 1 second ago = expired
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
