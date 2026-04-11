/**
 * core/src/state/migration.ts
 * Harness v2 JSON / JSONL → v3 SQLite migration script
 *
 * Imports v2 state files into the v3 SQLite DB.
 * Migration targets:
 *   .claude/state/session.json      → sessions table
 *   .claude/state/session.events.jsonl → signals table (task_completed, etc.)
 *   .claude/work-active.json         → work_states table
 *
 * Idempotent design: safe to re-run if migration is already complete.
 */
export interface MigrationResult {
    sessions: number;
    signals: number;
    workStates: number;
    skipped: boolean;
    errors: string[];
}
/**
 * Migrate v2 JSON/JSONL state files to the v3 SQLite DB.
 *
 * @param projectRoot - Project root path (default: process.cwd())
 * @param dbPath - SQLite DB path (default: <projectRoot>/.harness/state.db)
 * @returns Migration result
 */
export declare function migrate(projectRoot?: string, dbPath?: string): MigrationResult;
//# sourceMappingURL=migration.d.ts.map