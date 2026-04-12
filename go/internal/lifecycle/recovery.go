// Package lifecycle はエージェントのライフサイクル状態マシンを提供する。
package lifecycle

import "fmt"

// RecoveryLevel はリカバリの段階を表す型。
// SPEC.md §8 で定義された 4 段階に対応する。
type RecoveryLevel int

const (
	// SelfHeal は段階 1: 自己修復。エラー分析 → 自動修正 → リトライ。
	SelfHeal RecoveryLevel = iota
	// PeerHeal は段階 2: 仲間修復。別 Worker にタスクを委譲する。
	PeerHeal
	// LeadEscalation は段階 3: 指揮官介入。Lead セッションに escalation する。
	LeadEscalation
	// Abort は段階 4: 停止。ABORTED 状態へ遷移しユーザーに通知する。
	Abort
)

// String は RecoveryLevel の文字列表現を返す。
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

// RecoveryAction はリカバリの段階と実行すべきアクションを表す。
type RecoveryAction struct {
	// Level は現在のリカバリ段階。
	Level RecoveryLevel
	// Retry は自己修復後に同一タスクをリトライすべき場合に true。
	Retry bool
	// DelegateToWorker は別の Worker にタスクを委譲すべき場合に true。
	DelegateToWorker bool
	// EscalateToLead は Lead セッションに escalation すべき場合に true。
	EscalateToLead bool
	// Stop はリカバリ不可能な状態で停止すべき場合に true。
	Stop bool
	// Error はリカバリのトリガーとなったエラー（情報目的）。
	Error error
}

// RecoveryManager は 4 段階リカバリロジックを管理する構造体。
// StateMachine と連携して RECOVERING / ABORTED 状態への遷移を制御する。
type RecoveryManager struct {
	// sm はライフサイクル状態マシンへの参照。
	sm *StateMachine
	// attempts は HandleFailure が呼ばれた累計回数（0 始まり）。
	attempts int
	// maxSelfHeal は自己修復を試みる最大回数。デフォルト 3。
	maxSelfHeal int
	// maxPeerHeal は仲間修復を試みる最大回数。デフォルト 1。
	maxPeerHeal int
}

// NewRecoveryManager はデフォルト設定で RecoveryManager を生成する。
// maxSelfHeal=3, maxPeerHeal=1 が初期値として使用される。
func NewRecoveryManager(sm *StateMachine) *RecoveryManager {
	return &RecoveryManager{
		sm:          sm,
		attempts:    0,
		maxSelfHeal: 3,
		maxPeerHeal: 1,
	}
}

// HandleFailure は失敗を受け取り、現在の試行回数に基づいてリカバリアクションを返す。
// 段階の判定ルール（attempts は 0 始まり）:
//
//	0, 1, 2 番目 (attempts < maxSelfHeal=3)    → SelfHeal (Retry=true)
//	3 番目     (attempts < maxSelfHeal+maxPeerHeal=4) → PeerHeal (DelegateToWorker=true)
//	4 番目     (attempts < maxSelfHeal+maxPeerHeal+1=5) → LeadEscalation (EscalateToLead=true)
//	5 番目以降  (それ以外)                          → Abort (Stop=true)
//
// StateMachine が FAILED 状態にある場合は RECOVERING へ遷移する。
// Abort 段階では RECOVERING → ABORTED へ遷移する。
func (rm *RecoveryManager) HandleFailure(err error) RecoveryAction {
	// StateMachine が FAILED なら RECOVERING へ遷移を試みる
	if rm.sm.Current() == StateFailed {
		_ = rm.sm.Transition(StateRecovering, fmt.Sprintf("recovery attempt %d: %v", rm.attempts+1, err))
	}

	action := rm.determineAction(err)
	rm.attempts++

	// Abort 段階では StateMachine を ABORTED へ遷移する
	if action.Stop && rm.sm.Current() == StateRecovering {
		_ = rm.sm.Transition(StateAborted, fmt.Sprintf("all recovery attempts exhausted: %v", err))
	}

	return action
}

// determineAction は attempts の現在値に基づいてリカバリアクションを決定する内部メソッド。
// attempts インクリメント前に呼ばれる想定。
func (rm *RecoveryManager) determineAction(err error) RecoveryAction {
	switch {
	case rm.attempts < rm.maxSelfHeal:
		// 段階 1: 自己修復 (SelfHeal)
		return RecoveryAction{
			Level: SelfHeal,
			Retry: true,
			Error: err,
		}
	case rm.attempts < rm.maxSelfHeal+rm.maxPeerHeal:
		// 段階 2: 仲間修復 (PeerHeal)
		return RecoveryAction{
			Level:            PeerHeal,
			DelegateToWorker: true,
			Error:            err,
		}
	case rm.attempts < rm.maxSelfHeal+rm.maxPeerHeal+1:
		// 段階 3: 指揮官介入 (LeadEscalation)
		return RecoveryAction{
			Level:          LeadEscalation,
			EscalateToLead: true,
			Error:          err,
		}
	default:
		// 段階 4: 停止 (Abort)
		return RecoveryAction{
			Level: Abort,
			Stop:  true,
			Error: err,
		}
	}
}

// Attempts は現在のリカバリ試行回数を返す。
func (rm *RecoveryManager) Attempts() int {
	return rm.attempts
}

// Reset はリカバリ試行回数をゼロにリセットする。
// タスクが正常に完了した後に呼び出す。
func (rm *RecoveryManager) Reset() {
	rm.attempts = 0
}
