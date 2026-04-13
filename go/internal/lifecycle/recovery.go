// Package lifecycle provides the agent lifecycle state machine.
package lifecycle

import "fmt"

// RecoveryLevel represents the stage of recovery.
// Corresponds to the 4 stages defined in SPEC.md §8.
type RecoveryLevel int

const (
	// SelfHeal is stage 1: self-repair. Error analysis → automatic fix → retry.
	SelfHeal RecoveryLevel = iota
	// PeerHeal is stage 2: peer repair. Delegate task to another Worker.
	PeerHeal
	// LeadEscalation is stage 3: lead intervention. Escalate to the Lead session.
	LeadEscalation
	// Abort is stage 4: stop. Transition to ABORTED state and notify the user.
	Abort
)

// String returns the string representation of the RecoveryLevel.
func (l RecoveryLevel) String() string {
	switch l {
	case SelfHeal:
		return "SelfHeal"
	case PeerHeal:
		return "PeerHeal"
	case LeadEscalation:
		return "LeadEscalation"
	case Abort:
		return "Abort"
	default:
		return fmt.Sprintf("RecoveryLevel(%d)", int(l))
	}
}

// RecoveryAction represents a recovery stage and the action to take.
type RecoveryAction struct {
	// Level is the current recovery stage.
	Level RecoveryLevel
	// Retry is true when the same task should be retried after self-repair.
	Retry bool
	// DelegateToWorker is true when the task should be delegated to another Worker.
	DelegateToWorker bool
	// EscalateToLead is true when escalation to the Lead session is required.
	EscalateToLead bool
	// Stop is true when the agent must stop due to an unrecoverable state.
	Stop bool
	// Error is the error that triggered recovery (for informational purposes).
	Error error
}

// RecoveryManager manages the 4-stage recovery logic.
// Works with the StateMachine to control transitions to RECOVERING / ABORTED.
type RecoveryManager struct {
	// sm is a reference to the lifecycle state machine.
	sm *StateMachine
	// attempts is the cumulative number of HandleFailure calls (0-based).
	attempts int
	// maxSelfHeal is the maximum number of self-repair attempts. Default: 3.
	maxSelfHeal int
	// maxPeerHeal is the maximum number of peer-repair attempts. Default: 1.
	maxPeerHeal int
}

// NewRecoveryManager creates a RecoveryManager with default settings.
// Initial values: maxSelfHeal=3, maxPeerHeal=1.
func NewRecoveryManager(sm *StateMachine) *RecoveryManager {
	return &RecoveryManager{
		sm:          sm,
		attempts:    0,
		maxSelfHeal: 3,
		maxPeerHeal: 1,
	}
}

// HandleFailure receives a failure and returns a recovery action based on the current attempt count.
// Stage determination rules (attempts is 0-based):
//
//	0, 1, 2 (attempts < maxSelfHeal=3)              → SelfHeal (Retry=true)
//	3       (attempts < maxSelfHeal+maxPeerHeal=4)   → PeerHeal (DelegateToWorker=true)
//	4       (attempts < maxSelfHeal+maxPeerHeal+1=5) → LeadEscalation (EscalateToLead=true)
//	5+      (otherwise)                              → Abort (Stop=true)
//
// Transitions the StateMachine to RECOVERING if currently in FAILED state.
// At the Abort stage, transitions RECOVERING → ABORTED.
func (rm *RecoveryManager) HandleFailure(err error) RecoveryAction {
	// Attempt transition to RECOVERING if StateMachine is in FAILED state
	if rm.sm.Current() == StateFailed {
		_ = rm.sm.Transition(StateRecovering, fmt.Sprintf("recovery attempt %d: %v", rm.attempts+1, err))
	}

	action := rm.determineAction(err)
	rm.attempts++

	// At the Abort stage, transition the StateMachine to ABORTED
	if action.Stop && rm.sm.Current() == StateRecovering {
		_ = rm.sm.Transition(StateAborted, fmt.Sprintf("all recovery attempts exhausted: %v", err))
	}

	return action
}

// determineAction is an internal method that determines the recovery action based on the current attempts value.
// Called before incrementing attempts.
func (rm *RecoveryManager) determineAction(err error) RecoveryAction {
	switch {
	case rm.attempts < rm.maxSelfHeal:
		// Stage 1: Self-repair (SelfHeal)
		return RecoveryAction{
			Level: SelfHeal,
			Retry: true,
			Error: err,
		}
	case rm.attempts < rm.maxSelfHeal+rm.maxPeerHeal:
		// Stage 2: Peer repair (PeerHeal)
		return RecoveryAction{
			Level:            PeerHeal,
			DelegateToWorker: true,
			Error:            err,
		}
	case rm.attempts < rm.maxSelfHeal+rm.maxPeerHeal+1:
		// Stage 3: Lead intervention (LeadEscalation)
		return RecoveryAction{
			Level:          LeadEscalation,
			EscalateToLead: true,
			Error:          err,
		}
	default:
		// Stage 4: Stop (Abort)
		return RecoveryAction{
			Level: Abort,
			Stop:  true,
			Error: err,
		}
	}
}

// Attempts returns the current recovery attempt count.
func (rm *RecoveryManager) Attempts() int {
	return rm.attempts
}

// Reset resets the recovery attempt count to zero.
// Call after a task completes successfully.
func (rm *RecoveryManager) Reset() {
	rm.attempts = 0
}
