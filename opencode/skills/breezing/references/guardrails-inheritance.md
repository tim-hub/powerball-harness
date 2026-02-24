# Guardrails Inheritance

Agent Teams の Teammates に継承されるガードレールと、Breezing が採用する制約の実現方式。

## 自動継承される要素

Agent Teams の Teammates は Lead の環境を自動的に継承する。

### 継承される項目（公式ドキュメント明記）

| 項目 | 継承 | 説明 |
|------|------|------|
| CLAUDE.md | ✅ 自動 | プロジェクト・ディレクトリ階層の全 CLAUDE.md |
| .claude/rules/*.md | ✅ 自動 | ルールファイル全般 |
| MCP サーバー設定 | ✅ 自動 | Codex, chrome-devtools 等 |
| Skills | ✅ 自動 | プロジェクトコンテキストとして読み込み |
| 権限設定 | ✅ 自動 | spawn 時の `mode` パラメータで制御（`bypassPermissions` 推奨） |
| **Agent Memory** | ✅ 自動 | エージェント定義の `memory` フィールドに基づく永続メモリ注入 |

### 継承されない項目（確認済み）

| 項目 | 継承 | 対策 |
|------|------|------|
| **Hooks (PreToolUse/PostToolUse)** | **❌** | **spawn prompt で制約を明示** |
| 会話コンテキスト | ❌ | spawn prompt で必要な情報を渡す |
| Skills の暗黙的ロード | ❌ | spawn prompt で明示的に指定 |
| TaskList の状態 | ✅ 共有 | Agent Teams の共有タスクリスト |

> **Agent Memory の継承について**: `memory` フィールドはエージェント定義 (`agents/*.md`) の
> フロントマターで設定される。`subagent_type` でエージェントを指定して spawn すると、
> そのエージェントの `memory` 設定に基づき MEMORY.md が自動注入される。
> Breezing では Implementer に `task-worker` (`memory: project`)、
> Reviewer に `code-reviewer` (`memory: project`) を使用。

> **Hook 未継承の確認方法**: Agent Teams Teammate をスポーンし、`.claude/state/breezing-role-probe.json` への Write を実行。PreToolUse hook (pretooluse-guard.sh) が発火せず breezing-session-roles.json が生成されないこと、PostToolUse hook (emit-agent-trace.js) が発火せず agent-trace.jsonl にエントリが追加されないことを確認済み (Claude Code 2.1.33)。
>
> 公式ドキュメントの継承リスト: "CLAUDE.md, MCP servers, and skills" — Hooks は含まれていない。

### Bedrock/Vertex/Foundry 環境での注意（CC 2.1.41+）

CC 2.1.41 未満では Agent Teams が Bedrock/Vertex/Foundry で誤ったモデル識別子を使用するバグがあった。
`/breezing` を Bedrock/Vertex/Foundry 環境で使用する場合は **CC 2.1.41 以上を強く推奨**。

### Agent model field fix（CC 2.1.47+）

CC 2.1.47 でカスタムエージェントの `model` フィールドが Teammate spawn 時に正しく継承されるバグが修正された。

- **修正前**: `agents/*.md` の `model: sonnet` 等の指定が Teammate spawn 時に無視される場合があった
- **修正後**: `model` フィールドの値が確実に反映される
- **影響**: `video-scene-generator.md`（`model: sonnet`）等、モデルを明示指定しているエージェントの動作が確実に

Bedrock/Vertex/Foundry での Agent Teams 動作は CC 2.1.45 でも別途修正済み（モデル ID 解決）。

### Worktree isolation（CC 2.1.49+）

CC 2.1.49 で Task tool に `isolation: "worktree"` パラメータが追加された。

```typescript
// Task tool spawn 時に worktree 分離を指定
{
  description: "...",
  prompt: "...",
  subagent_type: "task-worker",
  isolation: "worktree"
}
```

- **効果**: git worktree で完全分離された環境で Implementer が作業する
- **解決する問題**: 同一ファイルへの並列書き込みによるコンフリクト（既知の制限事項）
- **既存の手順**: `skills/parallel-workflows/references/setup-worktrees.md` に手動 worktree 設定手順あり
- **CC 2.1.49+ 推奨**: `isolation: "worktree"` により手動 worktree 管理が不要になる
- **CC 2.1.50+**: エージェント定義でも `isolation: worktree` を宣言的に指定可能

### WorktreeCreate/WorktreeRemove Hook（CC 2.1.50+）

CC 2.1.50 で `WorktreeCreate` と `WorktreeRemove` フックイベントが追加された。

| イベント | 発火タイミング | 用途 |
|---------|--------------|------|
| **WorktreeCreate** | Worktree 作成時 | VCS セットアップ（ブランチ設定、初期状態の準備） |
| **WorktreeRemove** | Worktree 削除時 | クリーンアップ（一時ファイル、ステート削除） |

- **現状**: Harness では未実装（手動 worktree セットアップを使用）
- **将来対応**: Breezing 並列ワークフローの自動セットアップ・クリーンアップに活用可能

### バックグラウンドエージェントの停止（CC 2.1.49+）

CC 2.1.49 で背景エージェントの停止方法が変更された。

- **CC 2.1.49+**: `Ctrl+F`（2回押しで確認）でバックグラウンドエージェントを停止
- **旧バージョン**: ESC キー（**非推奨、CC 2.1.49+ では動作しない**）

> `/breezing` で並列 Implementer を停止する必要がある場合、`Ctrl+F` を使用すること。

### Lead 側で発火する Teammate 関連 Hook（2.1.33+）

Teammate 内部で Hook は動かないが、Lead 側で以下のイベントが発火する:

| イベント | 発火タイミング | ペイロード | 用途 |
|---------|--------------|-----------|------|
| **TeammateIdle** | Teammate のターン終了時 | `teammate_name`, `team_name` | ライフサイクル追跡 |
| **TaskCompleted** | Teammate がタスク完了時 | `teammate_name`, `task_id`, `task_subject`, `task_description` | タスク完了タイムライン |

**含まれないデータ**: トークン消費量、ツール使用数、処理時間（2.1.33 で実測確認済み）。

> **活用**: `.claude/hooks/log-teammate-idle.sh` 等で JSONL に記録し、完了レポートに「誰がいつ何を完了したか」のタイムラインを自動生成できる。Teammate 別コスト分析は `/cost` コマンドで全体確認のみ。

## Teammate の権限モデル: bypassPermissions + Hooks

### なぜ bypassPermissions が必要か

Teammate はバックグラウンドで実行されるため、**権限プロンプトを表示できない**。
`bypassPermissions` を指定しないと、Bash 等のツールが `"prompts unavailable"` で auto-deny される。

```
spawn 時に mode: "bypassPermissions" を指定
  → 全ツールの権限チェックをスキップ
  → ただし PreToolUse hooks は権限システムの前に独立して発火
  → hooks が危険操作を選択的に deny
```

### 安全性の多層防御

| レイヤー | 仕組み | bypassPermissions での動作 |
|---------|--------|--------------------------|
| **PreToolUse hooks** | pretooluse-guard（Lead セッション） | ⚠️ Teammate に未継承（将来対応予定） |
| **エージェント定義** | `disallowedTools` で使用禁止ツール指定 | ✅ 機能する |
| **Task(agent_type) 制限** | `Task(task-worker)` 等でスポーン可能なエージェント種類を制限 (CC 2.1.33+) | ✅ 機能する |
| **spawn prompt 制約** | ロール別の行動制約を明記 | ✅ 機能する |
| **.claude/rules/** | test-quality.md, implementation-quality.md | ✅ Teammate に継承 |
| **Lead の監視** | git diff で違反検知、SendMessage で警告 | ✅ 機能する |

> **公式ドキュメント根拠**: "PreToolUse hooks run BEFORE the permission system, and the hook output
> can determine whether to approve or deny the tool call in place of the permission system."
> Hooks は権限システムとは独立したレイヤーであり、bypassPermissions の影響を受けない。

## Breezing のロール制約方式

### 方式の選択

| 方式 | 強制力 | 採用状況 |
|------|--------|---------|
| **spawn prompt による制約** | 中（LLM が遵守する前提） | ✅ **現在の方式** |
| **エージェント定義 disallowedTools** | 高（ツール自体が使用不可） | ✅ **補助的に使用** |
| Hook 強制 (pretooluse-guard.sh) | 高（ツール実行前にブロック） | ⏸️ 休眠（Hooks が Teammate に継承されないため） |

### spawn prompt 制約 の仕組み

```
Step 1: Lead が spawn prompt でロール制約を明記
  Reviewer spawn: "Write, Edit を絶対に使用しない。Read-only。"
  Implementer spawn: "owns で指定されたファイルのみ編集可能。git commit 禁止。"

Step 2: Teammate は CLAUDE.md + rules/ + spawn prompt を読み込み

Step 3: Teammate は prompt 指示に従って行動
  → Reviewer: Read, Grep, Glob のみ使用
  → Implementer: owns 範囲内のファイルのみ Write/Edit

Step 4: Lead が定期的に検証
  → git diff で Reviewer が書き込んでいないか確認
  → Implementer が owns 外を編集していないか確認
```

### ロール別の制約（prompt で指示）

| ロール | Write/Edit | Bash | git commit | 強制方式 |
|--------|-----------|------|-----------|---------|
| **reviewer** | 禁止 | 書き込み系禁止 | 禁止 | spawn prompt |
| **implementer** | owns 内のみ | git commit/push 禁止 | 禁止 | spawn prompt |
| **lead** | Phase B: delegate mode で禁止, Phase A/C: 可 | Phase B: delegate mode で制限, Phase A/C: 可 | Phase C のみ | Phase A/C: ユーザー権限維持, Phase B: delegate mode |

### Lead の監視義務

spawn prompt 制約は LLM の遵守に依存するため、Lead が定期的に検証する:

```
1. git diff --name-only で変更ファイルを確認
   → Reviewer のファイルがあれば違反

2. Implementer の変更が owns 範囲内か確認
   → owns 外への変更は即座に SendMessage で警告

3. git commit が Lead 以外から行われていないか確認
   → git log --author で検証
```

## pretooluse-guard.sh の Breezing Role Guard（休眠中）

### 概要

pretooluse-guard.sh には session_id ベースのロール強制コードが実装済みだが、
**Hooks が Teammate に継承されないため、現時点では機能しない。**

将来 Claude Code が Hooks の Teammate 継承をサポートした場合に自動的に有効化される。

### 実装済みの機能（休眠中）

| 関数 | 役割 | 状態 |
|------|------|------|
| `check_breezing_role()` | session_id → role の検索 | 休眠 |
| `try_register_breezing_role()` | ロール登録 Write の検出と session_id マッピング | 休眠 |

### 有効化条件

以下が全て満たされた場合に自動有効化:

```
1. Claude Code が Hooks の Teammate 継承をサポート（公式アナウンス）
2. PreToolUse hook が Teammate の Write/Edit で発火することを確認
3. session_id が Teammate ごとに一意であることを確認
```

> **休眠コードを削除しない理由**:
> Hooks 継承が将来サポートされた場合、spawn prompt の「ロールマーカー Write」指示だけで
> 自動的に Hook 強制が有効化される。コードは既にテスト済みで、他のガード（Codex Mode 等）に
> 影響を与えない。

## Harness 固有のガードレール

### pretooluse-guard.sh（Lead セッションで適用）

Lead 自身のセッションでは pretooluse-guard.sh が正常に動作する:

| ガード | 内容 | Lead に適用 | Teammate に適用 |
|--------|------|------------|----------------|
| 危険コマンド防止 | rm -rf, git push --force 等 | ✅ | ❌ (Hooks 未継承) |
| 機密ファイル保護 | .env, credentials.json 等 | ✅ | ❌ |
| ブランチ保護 | main/master への直接操作 | ✅ | ❌ |
| Breezing Role Guard | session_id ベースのロール強制 | ✅ (休眠) | ❌ |

> **Teammate の安全策**: Teammate には CLAUDE.md + rules/ が継承されるため、
> ルールファイル（test-quality.md, implementation-quality.md）の制約は prompt レベルで機能する。

### .claude/rules/ のルールファイル（Teammate に継承 ✅）

| ルールファイル | 内容 | 対象 |
|---------------|------|------|
| test-quality.md | テスト改ざん禁止パターン | Implementer |
| implementation-quality.md | 形骸化実装禁止パターン | Implementer |
| skill-editing.md | スキルファイル編集規則 | Implementer |
| changelog.md | CHANGELOG 記載ルール | Lead (完了ステージ) |

## spawn prompt に追加すべき指示

### Implementer 追加指示

```markdown
## 品質ガードレール

### 絶対禁止パターン
- テスト改ざん: it.skip(), アサーション削除, expect値の直接返却
- スタブ実装: return null, return [], TODO コメントのみ
- lint ルール緩和: eslint-disable, @ts-ignore の追加
- 型安全性の回避: any 型の使用, as unknown as T

### ファイル所有権ルール
- owns: で指定されたファイルのみ編集可能
- 他の Implementer の owns ファイルには触らない
- 競合を検出したら Lead に SendMessage で報告
- ⚠️ この制約はプロンプトベース（Hook 強制なし）

### コミット禁止
- Implementer は git commit を絶対に実行しない
- コミットは Lead が完了ステージで一括実行
```

### Reviewer 追加指示

```markdown
## レビューガードレール

### Read-only 制約（厳守）
- Write, Edit ツールを絶対に使用しない
- Bash で書き込みコマンド (>, tee, sed -i, mv, cp, rm) を使用しない
- コードの修正は行わない、指摘のみ
- 修正は Implementer の責務
- ⚠️ この制約はプロンプトベース（Hook 強制なし）— 自己規律で遵守

### 客観性の維持
- Implementer の実装を事前知識なしでレビュー
- 「動いているから OK」は判定基準にならない
- セキュリティ、パフォーマンス、品質の3観点は必須

### 判定基準の厳守
- APPROVE: Critical/Major findings が 0 件
- REQUEST CHANGES: Major findings が 1 件以上
- REJECT: Critical findings が 1 件以上
- STOP: ビルド/テスト不通過
```

## Agent Teams 固有の注意事項

### 1. Teammate 間のファイル競合

```
Agent Teams は Teammates 間のファイルロックを提供しない
→ owns: アノテーションと addBlockedBy で論理的に制御
→ Lead が git diff で競合を監視し、問題発生時に介入
```

### 2. Teammate の逸脱防止

```
Teammate が spawn prompt の制約を逸脱した場合:
  → Lead が git diff --name-only で検出
  → SendMessage で Teammate に警告
  → 必要なら shutdown_request で停止
  → 最悪の場合、git checkout で変更を revert
```

### 3. トークン制限

```
各 Teammate は独立したコンテキストウィンドウを持つ
→ 大量のファイルを読み込むとコンテキストが溢れる
→ Implementer は対象ファイルのみ読み込むよう制限
→ Reviewer は git diff のみで判断可能な範囲に限定
```

## ultrawork との比較

| 項目 | ultrawork | breezing |
|------|-----------|---------|
| ガードレール適用 | Lead 自身に Hook 適用 | Lead に Hook 適用、Teammate は prompt 制約 |
| ファイル所有権 | なし (単一実行者) | owns: アノテーション (prompt 制約) |
| 逸脱防止 | pretooluse-guard.sh | prompt 制約 + Lead の git diff 監視 |
| コミット権限 | Lead が実行 | Lead のみ (Teammate は prompt で禁止) |
| Read-only 制約 | なし | Reviewer は prompt で制約 |
