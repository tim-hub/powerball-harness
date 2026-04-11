/**
 * lifecycle.ts — Session lifecycle management
 *
 * Absorbs logic from legacy session-related skills (session / session-init /
 * session-control / session-state / session-memory).
 * Centrally manages session start, end, and state transitions.
 */
import type { SessionState, SessionMode, Signal } from "../types.js";
/** Session execution phase */
export type SessionPhase = "active" | "paused" | "completed" | "failed";
/** Session context */
export interface SessionContext {
    sessionId: string;
    startedAt: Date;
    phase: SessionPhase;
    state: SessionState;
    /** Recent agent-trace entries */
    recentFiles: string[];
}
/**
 * Session initialization.
 * Performs environment checks, task status assessment, and handoff confirmation.
 */
export declare function initSession(opts: {
    sessionId: string;
    projectRoot: string;
    mode?: SessionMode;
}): SessionContext;
/**
 * Transition a session phase.
 * Throws an Error on invalid transitions.
 */
export declare function transitionSession(ctx: SessionContext, next: SessionPhase): SessionContext;
/** Session end summary */
export interface SessionSummary {
    sessionId: string;
    duration: number;
    finalPhase: SessionPhase;
    signals: Signal[];
}
/**
 * Session finalization.
 * Called for both completion and failure.
 */
export declare function finalizeSession(ctx: SessionContext, signals?: Signal[]): SessionSummary;
/**
 * Fork the current session context.
 * Returns an independent copy with a new session ID.
 */
export declare function forkSession(parent: SessionContext, newSessionId: string): SessionContext;
/** Minimal information for session resumption */
export interface ResumeInfo {
    sessionId: string;
    projectRoot: string;
    mode: SessionMode;
    lastPhase: SessionPhase;
}
/**
 * Resume a past session.
 * If lastPhase is completed or failed, treat as a new session.
 */
export declare function resumeSession(info: ResumeInfo): SessionContext;
//# sourceMappingURL=lifecycle.d.ts.map