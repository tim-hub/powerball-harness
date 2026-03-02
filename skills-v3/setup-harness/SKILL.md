---
name: setup-harness
description: "Unified setup skill for Harness v3. Project init, tool setup, 2-agent config, memory setup, symlink management. Use when user mentions: setup, initialization, new project, CI setup, codex CLI setup, harness-mem, agent setup, symlinks, /setup-harness. Do NOT load for: implementation, code review, release, or planning."
description-ja: "Harness v3 統合セットアップスキル。プロジェクト初期化・ツール設定・2エージェント構成・メモリ設定・symlink管理。以下で起動: セットアップ、初期化、新規プロジェクト、CIセットアップ、codex CLIセットアップ、harness-mem、エージェント設定、symlink、/setup-harness。実装・レビュー・リリース・プランニングには使わない。"
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash"]
argument-hint: "[init|ci|codex|harness-mem|symlinks|agents|localize]"
---

# Setup Skill (v3)

Harness v3 の統合セットアップスキル。
以下の旧スキルを統合:

- `setup` — 統合セットアップハブ
- `harness-init` — プロジェクト初期化
- `harness-update` — Harness アップデート
- `maintenance` — ファイル整理・クリーンアップ

## Quick Reference

| サブコマンド | 動作 |
|------------|------|
| `/setup init` | 新規プロジェクト初期化（CLAUDE.md + Plans.md + hooks）|
| `/setup ci` | CI/CD パイプライン設定 |
| `/setup codex` | Codex CLI インストール・設定 |
| `/setup harness-mem` | claude-mem 統合・メモリ設定 |
| `/setup symlinks` | skills-v3/ → ミラー間 symlink 更新 |
| `/setup agents` | agents-v3/ エージェント設定 |
| `/setup localize` | CLAUDE.md ルールのローカライズ |

## サブコマンド詳細

### init — プロジェクト初期化

新規プロジェクトに Harness v3 を導入する。

**生成ファイル**:
```
project/
├── CLAUDE.md            # プロジェクト設定
├── Plans.md             # タスク管理（空テンプレート）
├── .claude/
│   ├── settings.json    # Claude Code 設定
│   └── hooks.json       # フック設定（v3 シム）
└── hooks/
    ├── pre-tool.sh      # 薄いシム（→ core/src/index.ts）
    └── post-tool.sh     # 薄いシム（→ core/src/index.ts）
```

**フロー**:
1. プロジェクト種別を検出（Node.js/Python/Go/Rust/その他）
2. 最小限の CLAUDE.md を生成
3. Plans.md テンプレートを生成
4. hooks.json を配置

### ci — CI/CD 設定

GitHub Actions ワークフローを設定する。

```yaml
# .github/workflows/ci.yml 生成例
name: CI
on:
  push:
    branches: [main]
  pull_request:
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm ci && npm test
```

### codex — Codex CLI 設定

```bash
# インストール確認
which codex || npm install -g @openai/codex

# タイムアウトコマンド確認（macOS）
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")
# macOS の場合: brew install coreutils
```

**使用パターン**:
```bash
$TIMEOUT 120 codex exec "$(cat /tmp/prompt.md)" 2>/dev/null
```

### harness-mem — メモリ設定

claude-mem 統合を設定する。

```bash
# メモリディレクトリ作成
mkdir -p .claude/agent-memory/claude-code-harness-worker
mkdir -p .claude/agent-memory/claude-code-harness-reviewer

# MEMORY.md テンプレート配置
cat > .claude/agent-memory/claude-code-harness-worker/MEMORY.md << 'EOF'
# Worker Agent Memory

## Project Context
[プロジェクト概要]

## Patterns
[学習パターン]
EOF
```

### symlinks — Symlink 管理

v3 ではミラー同期を rsync から symlink に変更。

```bash
# skills-v3/ → codex/.codex/skills/ symlink 作成
WORKTREE=/path/to/worktree
cd $WORKTREE/codex/.codex/skills
for skill in plan execute review release setup; do
  ln -sf ../../../skills-v3/$skill $skill 2>/dev/null || true
done

# opencode/skills/ symlink 作成
cd $WORKTREE/opencode/skills
for skill in plan execute review release setup; do
  ln -sf ../../skills-v3/$skill $skill 2>/dev/null || true
done
```

### agents — エージェント設定

agents-v3/ の3エージェント構成を設定する。

```
agents-v3/
├── worker.md      # 実装担当（task-worker + codex-implementer + error-recovery）
├── reviewer.md    # レビュー担当（code-reviewer + plan-critic）
└── scaffolder.md  # 足場担当（project-analyzer + scaffolder）
```

### localize — ルールローカライズ

`.claude/rules/` のルールを現プロジェクトに適応する。

```bash
# ルール一覧確認
ls .claude/rules/

# プロジェクト固有ルールの追加
cat >> .claude/rules/project-rules.md << 'EOF'
# Project-Specific Rules
[プロジェクト固有ルール]
EOF
```

## Maintenance — ファイル整理

定期メンテナンスタスク:

| タスク | コマンド |
|--------|---------|
| 古いログ削除 | `find .claude/logs -mtime +30 -delete` |
| Plans.md 圧縮 | 完了タスクをアーカイブセクションに移動 |
| 古いトレース削除 | `tail -1000 .claude/state/agent-trace.jsonl > /tmp/trace && mv /tmp/trace .claude/state/agent-trace.jsonl` |

## 関連スキル

- `plan` — セットアップ後にプロジェクト計画を作成
- `execute` — セットアップ後にタスクを実行
- `review` — セットアップ設定をレビュー
