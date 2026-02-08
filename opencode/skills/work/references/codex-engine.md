# Codex Engine

`/work --codex` で Codex CLI を使った実装委託モード。
Claude は PM として調整のみ行い、実装は Codex Worker に委譲。

## Overview

```
/work --codex [scope]
    │
    ├─ Claude (PM): タスク分析・分割・レビュー
    │
    ├─ 単タスク → Codex CLI 直接呼び出し
    │
    └─ 複数タスク → 並列 Codex Worker
          ├─ Worker A → worktree-a → タスク A
          ├─ Worker B → worktree-b → タスク B
          └─ Worker C → worktree-c → タスク C
```

## Claude の役割（PM モード）

`--codex` モード時、Claude は **PM（Project Manager）** として機能。

### 許可される操作

| 操作 | 許可 | 説明 |
|------|------|------|
| ファイル読み込み | ✅ | Read, Glob, Grep |
| Codex Worker 呼び出し | ✅ | `Bash (codex exec)` |
| レビューと判定 | ✅ | 品質ゲート、証跡検証 |
| Plans.md 更新 | ✅ | 状態マーカーの更新のみ |
| Edit/Write | ❌ | **禁止**（pretooluse-guard でブロック） |

### 初期化時の設定

`--codex` フラグ指定時、`work-active.json` に `codex_mode: true` を設定:

```json
{
  "active": true,
  "started_at": "2026-02-08T10:00:00Z",
  "strategy": "iteration",
  "codex_mode": true,
  "bypass_guards": ["rm_rf", "git_push"],
  "allowed_rm_paths": ["node_modules", "dist", ".cache"]
}
```

## CLI Execution

```bash
# プロンプトファイル生成（base-instructions + タスク内容）
cat <<'CODEX_PROMPT' > /tmp/codex-prompt.md
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
$TIMEOUT 180 codex exec "$(cat /tmp/codex-prompt.md)" 2>/dev/null
EXIT_CODE=$?

# タイムアウト判定
if [ $EXIT_CODE -eq 124 ]; then
  echo "TIMEOUT: Codex CLI timed out after 180s"
fi
```

## AGENTS_SUMMARY Compliance

Worker は実行開始時に以下を出力する必要がある:

```
AGENTS_SUMMARY: <1行要約> | HASH:<SHA256先頭8文字>
```

- 入力: AGENTS.md ファイル内容（BOM除去、全行LF正規化）
- アルゴリズム: SHA256、Hex小文字、先頭8文字
- 欠落時: 即失敗 → 手動対応

## Worktree 必要性判定

並列実行時、まず Worktree が必要かを判定:

| 条件 | Worktree | 理由 |
|------|----------|------|
| タスク 1 つのみ | ❌ 不要 | 並列の意味がない |
| 全タスクが順次依存 | ❌ 不要 | 結局直列実行になる |
| owns: が全て重複 | ❌ 不要 | 同じファイルを触るため並列不可 |
| 並列可能タスク 2+ & ファイル分離 | ✅ 使用 | Worktree の価値あり |

## Quality Gates

| Gate | コマンド | 失敗時 | 最大リトライ |
|------|---------|--------|------------|
| lint | `npm run lint` | 自動修正指示 | 3 回 |
| type-check | `tsc --noEmit` | 型エラー修正指示 | 3 回 |
| test | `npm test` | テスト修正指示 | 3 回 |
| tamper | パターン検出 | 即停止 | 0 |

### 改ざん検出パターン

| パターン | 検出方法 |
|---------|---------|
| `it.skip()`, `test.skip()` | diff で新規追加を検出 |
| アサーション削除 | diff で `expect(` 行の減少を検出 |
| `eslint-disable` 追加 | diff で新規追加を検出 |

## Error Handling

| エラー | 対応 | リトライ |
|--------|------|---------|
| AGENTS_SUMMARY 欠落 | 即失敗 | 0 |
| ハッシュ不一致 | 段階的指示で再呼び出し | 3 |
| lint 失敗 | 自動修正指示 | 3 |
| テスト失敗 | 修正指示 | 3 |
| 改ざん検出 | 即停止 | 0 |
| マージ競合 | ユーザー判断 | 0 |

## Prerequisites

1. **Codex CLI**: `which codex` でパスが表示されること
2. **Git worktree**: `git --version` >= 2.5.0 (並列時)

## Related

- [auto-iteration.md](auto-iteration.md) - 自動反復ロジック
- [parallel-execution.md](parallel-execution.md) - 並列実行戦略
