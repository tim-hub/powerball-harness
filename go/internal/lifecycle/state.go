// Package lifecycle はエージェントのライフサイクル状態マシンを提供する。
//
// SPEC.md §8 に定義された全遷移ルールを宣言的に実装する。
// 正常系: SPAWNING → RUNNING → REVIEWING → APPROVED → COMMITTED
// 異常系: FAILED / CANCELLED / STALE / RECOVERING / ABORTED を含む。
package lifecycle

import (
	"fmt"
	"sync"
)

// AgentState はエージェントの状態を表す文字列型。
type AgentState string

const (
	// StateSpawning はエージェント起動中を表す。
	StateSpawning AgentState = "SPAWNING"
	// StateRunning はエージェントがタスクを実行中であることを表す。
	StateRunning AgentState = "RUNNING"
	// StateReviewing はレビューフェーズ中であることを表す。
	StateReviewing AgentState = "REVIEWING"
	// StateApproved はレビューが承認されたことを表す。
	StateApproved AgentState = "APPROVED"
	// StateCommitted はコミット完了を表す終端状態。
	StateCommitted AgentState = "COMMITTED"
	// StateFailed はエラーによる失敗を表す。
	StateFailed AgentState = "FAILED"
	// StateCancelled はユーザー中断による停止を表す終端状態。
	StateCancelled AgentState = "CANCELLED"
	// StateStale は 24h 超過による自動停止を表す。
	StateStale AgentState = "STALE"
	// StateRecovering はリカバリ処理中であることを表す。
	StateRecovering AgentState = "RECOVERING"
	// StateAborted はリカバリ失敗により人間介入が必要な終端状態。
	StateAborted AgentState = "ABORTED"
)

// terminalStates は遷移不可の終端状態の集合。
var terminalStates = map[AgentState]struct{}{
	StateCommitted: {},
	StateAborted:   {},
	StateCancelled: {},
}

// IsTerminal は指定した状態が終端状態かどうかを返す。
// COMMITTED / ABORTED / CANCELLED が終端状態に該当する。
func IsTerminal(state AgentState) bool {
	_, ok := terminalStates[state]
	return ok
}

// transitionKey は遷移テーブルのキーとなる (From, To) ペア。
type transitionKey struct {
	From AgentState
	To   AgentState
}

// validTransitions は許可された全遷移ルールを宣言的に定義する。
// SPEC.md §8 の正常系・異常系を網羅する。
var validTransitions = map[transitionKey]struct{}{
	// 正常系
	{StateSpawning, StateRunning}:   {}, // 起動成功
	{StateRunning, StateReviewing}:  {}, // 実行完了 → レビュー
	{StateReviewing, StateApproved}: {}, // レビュー承認
	{StateApproved, StateCommitted}: {}, // コミット完了

	// 異常系: FAILED
	{StateSpawning, StateFailed}:  {}, // 起動失敗
	{StateRunning, StateFailed}:   {}, // 実行中エラー・3回リトライ超過
	{StateReviewing, StateFailed}: {}, // レビュー中エラー

	// 異常系: CANCELLED
	{StateRunning, StateCancelled}:   {}, // ユーザー中断 (Ctrl+C)
	{StateReviewing, StateCancelled}: {}, // ユーザー中断

	// 異常系: STALE (24h 超過)
	{StateRunning, StateStale}:   {}, // 実行中に 24h 超過
	{StateReviewing, StateStale}: {}, // レビュー中に 24h 超過

	// 異常系: RECOVERING
	{StateFailed, StateRecovering}: {}, // リカバリ開始

	// 異常系: リカバリ結果
	{StateRecovering, StateRunning}: {}, // リカバリ成功 → 再実行
	{StateRecovering, StateAborted}: {}, // リカバリ失敗 → 人間介入必要
}

// Transition は単一の状態遷移を記録する構造体。
type Transition struct {
	// From は遷移元の状態。
	From AgentState
	// To は遷移先の状態。
	To AgentState
	// Trigger は遷移を引き起こしたイベントの説明。
	Trigger string
}

// StateMachine はエージェントのライフサイクル状態マシン。
// ゴルーチンセーフ。
type StateMachine struct {
	mu      sync.RWMutex
	current AgentState
	history []Transition
}

// NewStateMachine は SPAWNING 状態から始まる新しい StateMachine を返す。
func NewStateMachine() *StateMachine {
	return &StateMachine{
		current: StateSpawning,
		history: make([]Transition, 0),
	}
}

// Current は現在の状態を返す。
func (sm *StateMachine) Current() AgentState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// CanTransition は現在の状態から指定した状態への遷移が許可されているかを返す。
func (sm *StateMachine) CanTransition(to AgentState) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.canTransitionLocked(to)
}

// canTransitionLocked はロックを取得済みの前提で遷移可能かを判定する内部メソッド。
func (sm *StateMachine) canTransitionLocked(to AgentState) bool {
	key := transitionKey{From: sm.current, To: to}
	_, ok := validTransitions[key]
	return ok
}

// Transition は現在の状態から to へ遷移する。
// 遷移が許可されていない場合は error を返す。
// trigger には遷移を引き起こしたイベントの説明を渡す（ログ・デバッグ用）。
func (sm *StateMachine) Transition(to AgentState, trigger string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.canTransitionLocked(to) {
		return fmt.Errorf(
			"lifecycle: invalid transition %s → %s (trigger: %q)",
			sm.current, to, trigger,
		)
	}

	t := Transition{
		From:    sm.current,
		To:      to,
		Trigger: trigger,
	}
	sm.history = append(sm.history, t)
	sm.current = to
	return nil
}

// History は記録された遷移履歴のコピーを返す。
// 呼び出し元がスライスを変更しても内部状態には影響しない。
func (sm *StateMachine) History() []Transition {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]Transition, len(sm.history))
	copy(result, sm.history)
	return result
}
