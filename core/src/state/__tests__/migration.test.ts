/**
 * core/src/state/__tests__/migration.test.ts
 * Unit tests for migration.ts
 *
 * Includes actual filesystem operations, so uses a tmp directory.
 */

import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { mkdirSync, writeFileSync, existsSync, rmSync } from "node:fs";
import { resolve, join } from "node:path";
import { tmpdir } from "node:os";
import { migrate } from "../migration.js";
import { HarnessStore } from "../store.js";

// ============================================================
// Test utilities
// ============================================================

/** Create a temporary directory and return its path */
function createTmpProject(): string {
  const dir = join(tmpdir(), `harness-migration-test-${Date.now()}-${Math.random().toString(36).slice(2)}`);
  mkdirSync(dir, { recursive: true });
  mkdirSync(join(dir, ".claude", "state"), { recursive: true });
  mkdirSync(join(dir, ".harness"), { recursive: true });
  return dir;
}

/** Delete the temporary directory after tests */
function cleanupTmpProject(dir: string): void {
  if (existsSync(dir)) {
    rmSync(dir, { recursive: true, force: true });
  }
}

// ============================================================
// Tests
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
  // Already-migrated check
  // ------------------------------------------------------------------

  describe("skip when already migrated", () => {
    it("returns skipped: true if already migrated", () => {
      // First migration
      const first = migrate(projectRoot, dbPath);
      expect(first.skipped).toBe(false);

      // Second run is skipped
      const second = migrate(projectRoot, dbPath);
      expect(second.skipped).toBe(true);
      expect(second.sessions).toBe(0);
      expect(second.signals).toBe(0);
    });
  });

  // ------------------------------------------------------------------
  // Migration with no state files (empty migration)
  // ------------------------------------------------------------------

  describe("empty migration", () => {
    it("completes with 0 records when no migration target files exist", () => {
      const result = migrate(projectRoot, dbPath);
      expect(result.skipped).toBe(false);
      expect(result.sessions).toBe(0);
      expect(result.signals).toBe(0);
      expect(result.workStates).toBe(0);
      expect(result.errors).toHaveLength(0);
    });
  });

  // ------------------------------------------------------------------
  // session.json migration
  // ------------------------------------------------------------------

  describe("session.json migration", () => {
    it("can migrate a session", () => {
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

      // Verify saved in SQLite
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

    it("can migrate Unix timestamp format started_at", () => {
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

    it("can migrate even without session_id (uses default ID)", () => {
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

    it("continues with sessions: 0 for invalid JSON session.json", () => {
      const sessionFile = resolve(projectRoot, ".claude", "state", "session.json");
      writeFileSync(sessionFile, "{ invalid json }");

      const result = migrate(projectRoot, dbPath);
      expect(result.sessions).toBe(0);
      // No errors but sessions is 0 (null due to JSON parse failure)
    });

    it("renames session.json to .v2.bak after migration", () => {
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
  // session.events.jsonl migration
  // ------------------------------------------------------------------

  describe("session.events.jsonl migration", () => {
    it("can migrate signal events", () => {
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

    it("completes with 0 records for empty JSONL file", () => {
      const eventsFile = resolve(projectRoot, ".claude", "state", "session.events.jsonl");
      writeFileSync(eventsFile, "");

      const result = migrate(projectRoot, dbPath);
      expect(result.signals).toBe(0);
    });

    it("converts unknown event types to a fallback signal", () => {
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
  // work-active.json migration
  // ------------------------------------------------------------------

  describe("work-active.json migration", () => {
    it("can migrate a work_state", () => {
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
  // Combined migration (all files present)
  // ------------------------------------------------------------------

  describe("combined migration", () => {
    it("can migrate session + events + work-active all together", () => {
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
