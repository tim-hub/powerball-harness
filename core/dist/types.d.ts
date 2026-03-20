/**
 * core/src/types.ts
 * Harness v3 共通型定義
 *
 * Claude Code Hooks の stdin/stdout JSON スキーマと
 * ガードレールエンジンの内部型を定義する。
 */
/** PreToolUse / PostToolUse フックへの入力 */
export interface HookInput {
    /** 実行されようとしているツール名（例: "Bash", "Write"） */
    tool_name: string;
    /** ツールへの入力パラメータ */
    tool_input: Record<string, unknown>;
    /** セッション ID（Claude Code が設定） */
    session_id?: string;
    /** 現在の作業ディレクトリ */
    cwd?: string;
    /** プラグインルートディレクトリ */
    plugin_root?: string;
}
/** フックが返すアクション */
export type HookDecision = "approve" | "deny" | "ask";
/** フックの出力（Claude Code Hooks プロトコル） */
export interface HookResult {
    /** 実行を許可するか拒否するか */
    decision: HookDecision;
    /** ユーザーへの説明メッセージ */
    reason?: string;
    /** Claude への追加コンテキスト（systemMessage） */
    systemMessage?: string;
}
/** ガードルールの評価コンテキスト */
export interface RuleContext {
    input: HookInput;
    projectRoot: string;
    workMode: boolean;
    codexMode: boolean;
    breezingRole: string | null;
}
/** 単一ガードルールの定義 */
export interface GuardRule {
    /** ルール識別子（ログ・デバッグ用） */
    id: string;
    /** このルールが適用されるツール名のパターン（正規表現） */
    toolPattern: RegExp;
    /** ルールを評価する関数。一致しなければ null を返す */
    evaluate: (ctx: RuleContext) => HookResult | null;
}
/** エージェント間で交換するシグナルの種類 */
export type SignalType = "task_completed" | "task_failed" | "teammate_idle" | "session_start" | "session_end" | "stop_failure" | "request_review";
/** エージェント間シグナル */
export interface Signal {
    type: SignalType;
    /** 送信元セッション ID */
    from_session_id: string;
    /** 宛先セッション ID（省略時はブロードキャスト） */
    to_session_id?: string;
    /** シグナルのペイロード */
    payload: Record<string, unknown>;
    /** 送信時刻（ISO 8601） */
    timestamp: string;
}
/** タスク失敗の重大度 */
export type FailureSeverity = "warning" | "error" | "critical";
/** タスク失敗イベント */
export interface TaskFailure {
    /** 失敗したタスクの識別子 */
    task_id: string;
    /** 失敗の重大度 */
    severity: FailureSeverity;
    /** 失敗の説明 */
    message: string;
    /** スタックトレースまたは詳細情報 */
    detail?: string;
    /** 失敗時刻（ISO 8601） */
    timestamp: string;
    /** 試行回数 */
    attempt: number;
}
/** セッションの実行モード */
export type SessionMode = "normal" | "work" | "codex" | "breezing";
/** セッション状態 */
export interface SessionState {
    session_id: string;
    mode: SessionMode;
    project_root: string;
    started_at: string;
    /** work/breezing モードでのコンテキスト情報 */
    context?: Record<string, unknown>;
}
//# sourceMappingURL=types.d.ts.map