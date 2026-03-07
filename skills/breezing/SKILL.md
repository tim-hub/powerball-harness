---
name: breezing
description: "Agent Teams execution mode — backward-compatible alias for /harness-work with team orchestration. Trigger: breezing, team run, agent teams, run with team, 全部やって. Do NOT load for: single task, /work without team."
description-ja: "Agent Teams 実行モード — /harness-work のチーム協調エイリアス。breezing, チーム実行, 全部やって でトリガー。"
description-en: "Agent Teams execution mode — backward-compatible alias for /harness-work with team orchestration."
allowed-tools: ["Agent", "Read", "Write", "Edit", "Bash", "Grep", "Glob", "TaskCreate", "TaskUpdate", "TaskList", "TaskGet", "TeamCreate", "TeamDelete", "SendMessage", "WebSearch", "WebFetch"]
argument-hint: "[all|N-M|--codex|--parallel N|--no-discuss]"
user-invocable: true
---

# Breezing — Agent Teams Execution Mode

> **後方互換エイリアス**: `/harness-work` を Agent Teams モードで実行します。

## Quick Reference

```bash
/breezing                        # スコープを聞いてから実行
/breezing all                    # Plans.md 全タスクを完走
/breezing 3-6                    # タスク3〜6を完走
/breezing --codex all            # Codex CLI で全タスク完走
/breezing --parallel 2 all       # 2並列で全タスク完走
/breezing --no-discuss all       # 計画議論スキップで全タスク完走
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | 全未完了タスクを対象 | - |
| `N` or `N-M` | タスク番号/範囲指定 | - |
| `--codex` | Codex CLI で実装委託 | false |
| `--parallel N` | Implementer 並列数 | auto |
| `--no-commit` | 自動コミット抑制 | false |
| `--no-discuss` | 計画議論スキップ | false |

## Execution

**このスキルは `/harness-work` に委譲します。** 以下の設定で `/harness-work` を実行してください:

1. **引数をそのまま `/harness-work` に渡す**
2. **Agent Teams モードを強制** — TeamCreate → Worker spawn → Reviewer spawn の三者分離
3. **Lead は delegate 専念** — コードを直接書かない

### `/harness-work` との違い

| 特徴 | `/harness-work` | `/breezing` (このスキル) |
|------|-----------------|------------------------|
| 並列手段 | Task tool (サブエージェント) | **Agent Teams (Teammates)** |
| Lead の役割 | 調整+実装 | **delegate (調整専念)** |
| レビュー | Lead 自己レビュー | **独立 Reviewer Teammate** |
| デフォルトスコープ | 次のタスク | **全部** |

### Team Composition

| Role | Agent Type | Mode | 責務 |
|------|-----------|------|------|
| Lead | (self) | - | 調整・指揮・タスク分配 |
| Worker ×N | `general-purpose` | `bypassPermissions` | 実装 |
| Reviewer | `general-purpose` | `bypassPermissions` | 独立レビュー |

### Codex Mode (`--codex`)

Codex CLI にすべての実装を委託するモード:

```bash
# プロンプトは stdin パイプで渡す（ARG_MAX 対策）
CODEX_PROMPT=$(mktemp /tmp/codex-prompt-XXXXXX.md)
# タスク内容を書き出し
cat "$CODEX_PROMPT" | $TIMEOUT 120 codex exec - -a never -s workspace-write 2>>/tmp/harness-codex-$$.log
rm -f "$CODEX_PROMPT"
```

## Flow Summary

```
/breezing [scope] [--codex] [--parallel N] [--no-discuss]
    │
    ↓ Load /harness-work skill with Agent Teams mode
    │
Phase 0: Planning Discussion (--no-discuss でスキップ)
Phase A: Pre-delegate（チーム初期化）
Phase B: Delegate（Worker 実装 + Reviewer レビュー）
Phase C: Post-delegate（統合検証 + Plans.md 更新 + commit）
```

## Active Monitoring with /loop (v2.1.71+)

`/loop` コマンドで Cron 風の定期実行が可能。Breezing セッション中のタスク進捗監視に活用する。

```bash
/loop 5m /sync-status    # 5分ごとにタスク進捗をチェック
/loop 10m check stale    # 10分ごとに停滞タスクを検出
```

### TeammateIdle との使い分け

| 方式 | 発火タイミング | 用途 |
|------|---------------|------|
| `TeammateIdle` hook | Teammate がアイドル状態になった時（受動的） | 次タスクの即時割り当て |
| `/loop` | 指定間隔で定期実行（能動的） | 全体進捗の俯瞰・停滞検出 |

TeammateIdle は個別 Teammate の空き検出、`/loop` はチーム全体の定期ヘルスチェックとして併用する。

## Background Agent (v2.1.71+)

v2.1.71 で出力パスが完了通知に含まれるようになり、Background Agent が安全に利用可能になった。

### 使用例

長時間実行タスク（大規模リファクタリング、全ファイルマイグレーション等）を Background Agent に委任:

```
Task tool で run_in_background: true を指定
→ エージェントがバックグラウンドで実行
→ 完了通知に出力ファイルパスが含まれる
→ 圧縮後でも結果を回収可能
```

### 注意事項

- Background Agent はコンテキスト圧縮の影響を受けるため、結果は出力パス経由で回収する
- 短時間で完了するタスクには通常の Task tool（フォアグラウンド）を推奨
- `/loop` と組み合わせて Background Agent の完了状態を定期チェックすると効果的

## Related Skills

- `/harness-work` — 単一タスクからチーム実行まで（本体）
- `/harness-review` — コードレビュー（breezing 内で自動起動）
