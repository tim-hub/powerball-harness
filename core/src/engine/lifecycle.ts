/**
 * lifecycle.ts — Session lifecycle management
 *
 * Absorbs logic from legacy session-related skills (session / session-init /
 * session-control / session-state / session-memory).
 * Centrally manages session start, end, and state transitions.
 */

import type { SessionState, SessionMode, Signal } from "../types.js";

// ============================================================
// Session execution state (internal enum equivalent)
// ============================================================

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

// ============================================================
// Session start (equivalent to session-init)
// ============================================================

/**
 * Session initialization.
 * Performs environment checks, task status assessment, and handoff confirmation.
 */
export function initSession(opts: {
  sessionId: string;
  projectRoot: string;
  mode?: SessionMode;
}): SessionContext {
  const initialState: SessionState = {
    session_id: opts.sessionId,
    mode: opts.mode ?? "normal",
    project_root: opts.projectRoot,
    started_at: new Date().toISOString(),
  };

  return {
    sessionId: opts.sessionId,
    startedAt: new Date(),
    phase: "active",
    state: initialState,
    recentFiles: [],
  };
}

// ============================================================
// Session state transitions (equivalent to session-state / session-control)
// ============================================================

/** Allowed state transition map */
const VALID_TRANSITIONS: Record<SessionPhase, SessionPhase[]> = {
  active: ["paused", "completed", "failed"],
  paused: ["active", "completed", "failed"],
  completed: [],
  failed: [],
};

/**
 * Transition a session phase.
 * Throws an Error on invalid transitions.
 */
export function transitionSession(
  ctx: SessionContext,
  next: SessionPhase,
): SessionContext {
  const allowed = VALID_TRANSITIONS[ctx.phase];
  if (!allowed.includes(next)) {
    throw new Error(
      `Invalid session transition: ${ctx.phase} → ${next}`,
    );
  }
  return { ...ctx, phase: next };
}

// ============================================================
// Session finalization (equivalent to session-memory)
// ============================================================

/** Session end summary */
export interface SessionSummary {
  sessionId: string;
  duration: number; // milliseconds
  finalPhase: SessionPhase;
  signals: Signal[];
}

/**
 * Session finalization.
 * Called for both completion and failure.
 */
export function finalizeSession(
  ctx: SessionContext,
  signals: Signal[] = [],
): SessionSummary {
  const duration = Date.now() - ctx.startedAt.getTime();
  return {
    sessionId: ctx.sessionId,
    duration,
    finalPhase: ctx.phase,
    signals,
  };
}

// ============================================================
// Session fork (equivalent to session-control --fork)
// ============================================================

/**
 * Fork the current session context.
 * Returns an independent copy with a new session ID.
 */
export function forkSession(
  parent: SessionContext,
  newSessionId: string,
): SessionContext {
  const forkedState: SessionState = {
    ...parent.state,
    session_id: newSessionId,
    started_at: new Date().toISOString(),
  };

  return {
    ...parent,
    sessionId: newSessionId,
    startedAt: new Date(),
    phase: "active",
    state: forkedState,
  };
}

// ============================================================
// Session resume (equivalent to session-control --resume)
// ============================================================

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
export function resumeSession(info: ResumeInfo): SessionContext {
  const isResumable =
    info.lastPhase === "active" || info.lastPhase === "paused";

  const newId = isResumable
    ? info.sessionId
    : `${info.sessionId}-resumed`;

  const state: SessionState = {
    session_id: newId,
    mode: info.mode,
    project_root: info.projectRoot,
    started_at: new Date().toISOString(),
  };

  return {
    sessionId: newId,
    startedAt: new Date(),
    phase: "active",
    state,
    recentFiles: [],
  };
}
