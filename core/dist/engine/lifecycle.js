/**
 * lifecycle.ts — Session lifecycle management
 *
 * Absorbs logic from legacy session-related skills (session / session-init /
 * session-control / session-state / session-memory).
 * Centrally manages session start, end, and state transitions.
 */
// ============================================================
// Session start (equivalent to session-init)
// ============================================================
/**
 * Session initialization.
 * Performs environment checks, task status assessment, and handoff confirmation.
 */
export function initSession(opts) {
    const initialState = {
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
const VALID_TRANSITIONS = {
    active: ["paused", "completed", "failed"],
    paused: ["active", "completed", "failed"],
    completed: [],
    failed: [],
};
/**
 * Transition a session phase.
 * Throws an Error on invalid transitions.
 */
export function transitionSession(ctx, next) {
    const allowed = VALID_TRANSITIONS[ctx.phase];
    if (!allowed.includes(next)) {
        throw new Error(`Invalid session transition: ${ctx.phase} → ${next}`);
    }
    return { ...ctx, phase: next };
}
/**
 * Session finalization.
 * Called for both completion and failure.
 */
export function finalizeSession(ctx, signals = []) {
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
export function forkSession(parent, newSessionId) {
    const forkedState = {
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
/**
 * Resume a past session.
 * If lastPhase is completed or failed, treat as a new session.
 */
export function resumeSession(info) {
    const isResumable = info.lastPhase === "active" || info.lastPhase === "paused";
    const newId = isResumable
        ? info.sessionId
        : `${info.sessionId}-resumed`;
    const state = {
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
//# sourceMappingURL=lifecycle.js.map