# Codex Engine (Breezing)

`/breezing --codex` で Codex CLI にすべての実装を委託するモード。
Claude は Lead（指揮）と Reviewer（品質保証）に専念する。

## Overview

```
/breezing --codex [scope] [--parallel N]
    │
Phase A: Pre-delegate（準備 + breezing-active.json 書き込み + Team spawn）
    │
    ↓ delegate mode ON
Phase B: Delegate
Lead ─ 指揮のみ（TaskCreate/TaskUpdate/SendMessage のみ）
  │
  ├── Codex Implementer #1 (sonnet) ─ Codex CLI 呼び出し + Quality Gates
  ├── Codex Implementer #2 (sonnet) ─ 同上 (独立タスク)
  ├── [Codex Implementer #3] (sonnet) ─ 同上 (必要に応じて)
  │
  └── Reviewer (sonnet) ─ harness-review 4 観点 + 判定
    │
    ↓ delegate mode OFF
Phase C: Post-delegate（Plans.md 更新 + git commit + cleanup）
```

## Compaction Recovery

**Compaction が発生した場合（コンテキストが圧縮された場合）の復元手順:**

1. `.claude/state/breezing-active.json` を Read する（ファイルが存在しない/読めない場合は停止してユーザーに確認）
2. `impl_mode` が `"codex"` であることを確認
3. `team_name` で TaskList が存在するか確認（`~/.claude/tasks/{team_name}/`）
4. Team が消失していれば再作成（Phase A として実行 — delegate mode に入る前に spawn を完了）:
   - TeamCreate
   - codex-implementer Teammate を `team.implementer_count` 個 spawn（`mode: "bypassPermissions"`）
   - code-reviewer Teammate を spawn（`mode: "bypassPermissions"`）
   - delegate mode ON → Phase B へ遷移
5. `team_name` がまだない（準備ステージ中の compaction）場合は、準備ステージの範囲確認から再開
6. TaskList で未完了タスクを確認し、サイクルを再開

**絶対禁止**: breezing-active.json に `impl_mode: "codex"` がある限り、Lead が Write/Edit でリポジトリのソースコードを直接書くことは禁止。
（Lead が編集してよいもの: breezing-active.json, Plans.md のマーカー更新のみ）

## breezing-active.json Schema (--codex 版)

```jsonc
{
  "session_id": "breezing-codex-20260208-0300",
  "started_at": "2026-02-08T03:00:00Z",
  "team_name": "breezing-auth-feature",
  "task_range": "認証機能からユーザー管理まで",
  "impl_mode": "codex",
  "plans_md_mapping": {
    "task-1": "4.1",
    "task-2": "4.2",
    "task-3": "4.3"
  },
  "options": {
    "parallel": 2
  },
  "team": {
    "implementer_count": 2,
    "reviewer_count": 1,
    "model": "sonnet"
  },
  "isolation": {
    "strategy": "worktree",
    "worktree_base": "../worktrees"
  },
  "review": {
    "retake_count": 0,
    "max_retakes": 3
  }
}
```

## Team Composition

### Lead (自分自身)

| 項目 | 設定 |
|------|------|
| **モード** | Phase A/C: ユーザーのパーミッションモード維持, Phase B: delegate mode |
| **責務** | タスク分配、進捗監視、リテイク分解、ファイル分離戦略決定 |
| **Phase A ツール** | Write, Edit, Bash, TaskCreate（準備・初期化用） |
| **Phase B ツール** | TaskCreate, TaskUpdate, TaskList, TaskGet, SendMessage のみ |
| **Phase C ツール** | Write, Edit, Bash（Plans.md 更新、git commit、cleanup 用） |
| **禁止事項** | Phase B 中の Write, Edit, Bash による直接実装 |

### Codex Implementer

| 項目 | 設定 |
|------|------|
| **subagent_type** | `claude-code-harness:codex-implementer` |
| **数** | 1〜3 (独立タスク数に基づく自動決定) |
| **責務** | Codex CLI 呼び出し、AGENTS_SUMMARY 検証、Quality Gates 実行 |
| **ツール** | `Bash (codex exec)` |

### Reviewer

| 項目 | 設定 |
|------|------|
| **subagent_type** | `claude-code-harness:code-reviewer` |
| **数** | 1 (常に) |
| **制約** | Read-only (Write/Edit 禁止) |

## Codex Implementer Flow

```
1. TaskList で pending かつ blockedBy が空のタスクを確認
2. 最も ID が小さいタスクを self-claim (TaskUpdate → in_progress)
3. codex-implementer フロー実行:
   a. base-instructions 生成（.claude/rules/*.md 連結 + AGENTS.md 準拠指示 + owns 制約）
   b. (worktree モード時) git worktree 準備
   c. codex exec (CLI) 呼び出し
   d. AGENTS_SUMMARY 検証
   e. Quality Gates (lint → type-check → test)
   f. (worktree モード時) cherry-pick + worktree 削除
4. 成功 → TaskUpdate(completed) → 次タスクへ
5. 3回失敗 → Lead に SendMessage でエスカレーション
6. 残りタスクなし → Lead に完了報告 (SendMessage)
```

## CLI Execution

```bash
# プロンプトファイル生成（base-instructions + タスク内容）
cat <<'CODEX_PROMPT' > /tmp/codex-prompt-{id}.md
## プロジェクトルール
{.claude/rules/*.md 連結}

## 必須: AGENTS.md 準拠
{AGENTS_SUMMARY 証跡出力指示}

---
{タスク内容}
CODEX_PROMPT

# タイムアウトコマンド検出（macOS: brew install coreutils）
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")

# 実行（タイムアウト 180秒）
$TIMEOUT 180 codex exec "$(cat /tmp/codex-prompt-{id}.md)" 2>/dev/null
EXIT_CODE=$?

# タイムアウト判定
if [ $EXIT_CODE -eq 124 ]; then
  echo "TIMEOUT: Codex CLI timed out after 180s"
fi
```

## AGENTS_SUMMARY Compliance

Worker は実行開始時に以下を出力:

```
AGENTS_SUMMARY: <1行要約> | HASH:<SHA256先頭8文字>
```

| 結果 | アクション | リトライ |
|------|-----------|---------|
| 証跡あり + ハッシュ一致 | Quality Gates へ | - |
| 証跡あり + ハッシュ不一致 | リトライ（指示を明確化） | 最大 3 回 |
| 証跡欠落 | 即失敗 → Lead にエスカレーション | なし |

## Quality Gates

| Gate | コマンド | 失敗時 | 最大リトライ |
|------|---------|--------|------------|
| lint | `npm run lint` | 自動修正指示 | 3 回 |
| type-check | `tsc --noEmit` | 型エラー修正指示 | 3 回 |
| test | `npm test` | テスト修正指示 | 3 回 |
| tamper | パターン検出 | 即停止 | 0 |

## Prerequisites

1. **Agent Teams 有効化**: `settings.json` に `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
2. **Codex CLI インストール済み**: `which codex` でパスが表示されること
3. **Plans.md** が存在し、未完了タスクがあること

Codex なしで実行する場合は `--codex` を外して `/breezing` を使用。

## Completion Report (--codex 版)

```markdown
🏇 Breezing Complete! (Codex Engine)

## Summary
- 対象: 認証機能からユーザー管理まで (3 タスク)
- 所要時間: 12 分
- Codex Implementer: 2 並列
- 分離戦略: worktree
- リテイク: 1 回

## AGENTS_SUMMARY Compliance
- 全タスク検証通過: 3/3

## Review
- 判定: APPROVE (Grade: A)

Codex が手を動かし、Claude がチェック。楽勝でした 🐎
```

## Related

- [team-composition.md](team-composition.md) - 標準チーム構成
- [execution-flow.md](execution-flow.md) - 実行フロー詳細
- [review-retake-loop.md](review-retake-loop.md) - リテイクループ
