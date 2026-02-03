# Plans.md - Claude Code Permissions 改善プラン

## 概要

Claude Code の Permissions ドキュメントに基づき、Harness プロジェクトの設定を改善する。

## 背景

Claude Code の Permissions が独立ページ化され、以下の新機能・推奨設定が明確化：
- Permission 構文の正規化（`:*` → ` *`）
- Sandbox によるセキュリティ強化
- 新フックイベント（PreCompact, SessionEnd）
- Agent 型フック
- Skills の `context: fork` と プリロード

---

## Phase 1: 分析・調査（リサーチ）

### 1.1 Sandbox 影響分析 `[research]`

| Task | 内容 | Status |
|------|------|--------|
| 1.1.1 | 現在の bypassPermissions 使用状況の確認 | `cc:TODO` |
| 1.1.2 | Sandbox 有効時の操作感への影響調査 | `cc:TODO` |
| 1.1.3 | excludedCommands に含めるべきコマンドの特定 | `cc:TODO` |
| 1.1.4 | allowedDomains の候補リスト作成 | `cc:TODO` |

**成果物**: Sandbox 設定推奨案

---

### 1.2 PreCompact フック活用分析 `[research]`

| Task | 内容 | Status |
|------|------|--------|
| 1.2.1 | Harness の既存 context 保存機能の確認 | `cc:TODO` |
| 1.2.2 | compaction 時に失われる情報の特定 | `cc:TODO` |
| 1.2.3 | PreCompact で保存すべき情報の設計 | `cc:TODO` |
| 1.2.4 | session-resume との連携方法の検討 | `cc:TODO` |

**成果物**: PreCompact フック実装提案

---

### 1.3 Agent 型フック活用分析 `[research]`

| Task | 内容 | Status |
|------|------|--------|
| 1.3.1 | 現在の Stop フック（prompt 型）の評価 | `cc:TODO` |
| 1.3.2 | Agent 型に切り替えるメリット・デメリット整理 | `cc:TODO` |
| 1.3.3 | Agent 型が適切なユースケースの特定 | `cc:TODO` |
| 1.3.4 | タイムアウト・コスト影響の検討 | `cc:TODO` |

**成果物**: Agent 型フック採用判断と設計

---

### 1.4 context: fork 適用分析 `[research]`

| Task | 内容 | Status |
|------|------|--------|
| 1.4.1 | 全スキル（45個）の分類確認 | `cc:DONE` |
| 1.4.2 | 既に fork を使用しているスキル（4個）の確認 | `cc:DONE` |
| 1.4.3 | fork 追加候補スキルの選定 | `cc:TODO` |
| 1.4.4 | fork 不要スキルの理由整理 | `cc:TODO` |

**調査結果**:
- 既存 fork スキル: `agent-browser`, `ci`, `harness-review`, `troubleshoot`
- 候補: 重い調査系、外部連携系

---

### 1.5 Attribution と AgentTrace 統合分析 `[research]`

| Task | 内容 | Status |
|------|------|--------|
| 1.5.1 | AgentTrace の現在のスキーマ確認 | `cc:DONE` |
| 1.5.2 | Attribution 情報の取得可能性調査 | `cc:DONE` |
| 1.5.3 | スキーマ拡張設計（v0.2.0） | `cc:TODO` |
| 1.5.4 | 環境変数からの情報取得方法確認 | `cc:TODO` |

**調査結果**:
- AgentTrace: v0.1.0、JSONL形式、PostToolUse で記録
- Attribution 追加可能（plugin.json から license, version 取得可）
- スキーマ拡張で `metadata.attribution` フィールド追加が妥当

---

### 1.6 SessionEnd フック重複確認 `[research]`

| Task | 内容 | Status |
|------|------|--------|
| 1.6.1 | 既存 Stop フックとの機能比較 | `cc:DONE` |
| 1.6.2 | SessionEnd の発火条件確認 | `cc:DONE` |
| 1.6.3 | 重複の有無判定 | `cc:DONE` |

**調査結果**:
- SessionEnd フックは**未実装**（hooks.json に定義なし）
- Stop フックで session-summary を実行中
- Stop: 応答終了ごと / SessionEnd: セッション完全終了時
- **重複なし** - 異なるユースケースで併用可能

---

## Phase 2: 実装（確定タスク）

### 2.1 Permission 構文修正 `[feature:breaking-change]`

| Task | 内容 | Status |
|------|------|--------|
| 2.1.1 | `.claude/settings.json` の `:*` → ` *` 置換 | `cc:TODO` |
| 2.1.2 | `~/.claude/settings.json` の `:*` → ` *` 置換 | `cc:TODO` |
| 2.1.3 | `.claude/settings.local.json` の `:*` → ` *` 置換 | `cc:TODO` |
| 2.1.4 | 動作確認（git, npm, test コマンド） | `cc:TODO` |

**影響**: 非推奨構文の修正。将来の互換性確保。

---

### 2.2 Skills プリロード導入 `[feature]`

| Task | 内容 | Status |
|------|------|--------|
| 2.2.1 | 既存エージェントの skills フィールド確認 | `cc:DONE` |
| 2.2.2 | プリロード最適化の余地がないか検討 | `cc:TODO` |
| 2.2.3 | 新規エージェント作成時のガイドライン文書化 | `cc:TODO` |

**調査結果**:
- 全9エージェントで既に skills プリロード使用中
- 設計は既に最適化済み
- ドキュメント整備が主なタスク

---

## Phase 3: 実装（分析結果に基づく）

### 3.1 Sandbox 設定導入（条件付き）

| Task | 内容 | Status |
|------|------|--------|
| 3.1.1 | 推奨設定の作成 | `cc:TODO` |
| 3.1.2 | ユーザー設定への適用 | `cc:TODO` |
| 3.1.3 | 動作確認とフォールバック設計 | `cc:TODO` |

---

### 3.2 PreCompact フック実装

| Task | 内容 | Status |
|------|------|--------|
| 3.2.1 | pre-compact-save スクリプト作成 | `cc:TODO` |
| 3.2.2 | hooks.json への追加 | `cc:TODO` |
| 3.2.3 | session-resume との連携テスト | `cc:TODO` |

---

### 3.3 Agent 型フック検討（オプション）

| Task | 内容 | Status |
|------|------|--------|
| 3.3.1 | Stop フックの Agent 型移行判断 | `cc:TODO` |
| 3.3.2 | 実装（必要な場合のみ） | `cc:TODO` |

---

### 3.4 context: fork 追加適用

| Task | 内容 | Status |
|------|------|--------|
| 3.4.1 | 候補スキルへの fork 適用 | `cc:TODO` |
| 3.4.2 | 動作確認とパフォーマンス検証 | `cc:TODO` |

---

### 3.5 Attribution 統合

| Task | 内容 | Status |
|------|------|--------|
| 3.5.1 | AgentTrace スキーマ v0.2.0 作成 | `cc:TODO` |
| 3.5.2 | emit-agent-trace.js 拡張 | `cc:TODO` |
| 3.5.3 | テスト追加 | `cc:TODO` |

---

### 3.6 SessionEnd フック導入

| Task | 内容 | Status |
|------|------|--------|
| 3.6.1 | SessionEnd の用途設計 | `cc:TODO` |
| 3.6.2 | hooks.json への追加 | `cc:TODO` |
| 3.6.3 | 一時ファイルクリーンアップ実装 | `cc:TODO` |

---

## 完了基準

- [ ] Phase 1: 全分析タスク完了
- [ ] Phase 2: 確定タスク実装完了
- [ ] Phase 3: 分析結果に基づく実装完了
- [ ] 動作確認・テスト合格
- [ ] CHANGELOG 更新
- [ ] バージョンアップ

## 技術決定事項

| 項目 | 決定 | 根拠 |
|------|------|------|
| Permission 構文 | ` *` に統一 | 公式非推奨対応 |
| SessionEnd | Stop と併用 | 異なるユースケース |
| Skills プリロード | 既存設計を維持 | 既に最適化済み |
| AgentTrace Attribution | v0.2.0 で追加 | 後方互換性維持 |
