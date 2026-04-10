---
name: HAR:setup
description: "HAR:プロジェクト初期化・ツール設定・エージェント構成・メモリ設定・skill mirror 同期を担当。セットアップ、初期化、新規プロジェクト、CI/Codex CLI セットアップ、harness-mem、mirror で起動。実装・レビュー・リリース・プランニングには使わない。"
description-en: "HAR: Project init, tool setup, agent config, memory setup, skill mirror sync. Trigger: setup, init, new project, CI/Codex setup, harness-mem, mirror. Do NOT load for: implementation, review, release, planning."
description-ja: "HAR:プロジェクト初期化・ツール設定・エージェント構成・メモリ設定・skill mirror 同期を担当。セットアップ、初期化、新規プロジェクト、CI/Codex CLI セットアップ、harness-mem、mirror で起動。実装・レビュー・リリース・プランニングには使わない。"
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash"]
argument-hint: "[init|ci|codex|harness-mem|mirrors|agents|localize]"
effort: medium
---

# Harness Setup

Harness の統合セットアップスキル。
以下の旧スキルを統合:

- `setup` — 統合セットアップハブ
- `harness-init` — プロジェクト初期化
- `harness-update` — Harness アップデート
- `maintenance` — ファイル整理・クリーンアップ

## Quick Reference

| サブコマンド | 動作 |
|------------|------|
| `harness-setup init` | 新規プロジェクト初期化（CLAUDE.md + Plans.md + hooks + sync + doctor）|
| `harness-setup ci` | CI/CD パイプライン設定 |
| `harness-setup codex` | Codex CLI インストール・設定 |
| `harness-setup harness-mem` | harness-mem 統合・メモリ設定 |
| `harness-setup mirrors` | skills/ → 公開 mirror bundle 更新 |
| `harness-setup agents` | agents/ エージェント設定 |
| `harness-setup localize` | CLAUDE.md ルールのローカライズ |

## サブコマンド詳細

### init — プロジェクト初期化

新規プロジェクトに Harness を導入する。

**生成ファイル**:
```
project/
├── CLAUDE.md            # プロジェクト設定
├── Plans.md             # タスク管理（空テンプレート）
├── .claude/
│   ├── settings.json    # Claude Code 設定
│   └── hooks.json       # フック設定（Go バイナリ）
└── hooks/
    ├── pre-tool.sh      # 薄いシム（→ core/src/index.ts）
    └── post-tool.sh     # 薄いシム（→ core/src/index.ts）
```

**フロー**:
1. プロジェクト種別を検出（Node.js/Python/Go/Rust/その他）
2. 最小限の CLAUDE.md を生成
3. Plans.md テンプレートを生成
4. hooks.json を配置
5. **Go バイナリ検証**: `harness version` でバイナリが利用可能か確認（v4.0 以降 Node.js 不要）
6. **プラグインファイル同期**: `harness sync` で `.claude-plugin/` 配下のファイルを最新に同期
7. **ヘルスチェック**: `harness doctor` で全チェック項目をパス。問題があれば修正案を提示

### Go バイナリ検証

```bash
# バイナリの存在と動作を確認
harness version
# 例: harness v4.0.0 (go1.22.0, darwin/arm64)
```

v4.0 以降、Harness のコアエンジンは Go バイナリに移行した。
Node.js は不要。バイナリは `bin/harness`（または PATH 上の `harness`）を使用する。

### プラグインファイル同期

```bash
# .claude-plugin/ 配下のファイルを最新に同期
harness sync

# 同期内容の確認のみ（変更なし）
harness sync --dry-run
```

`harness sync` は skills/ の SSOT から各 mirror（codex/.codex/skills/、opencode/skills/）へ
変更を伝播させる。init 後に必ず実行すること。

### ヘルスチェック

```bash
# 全チェック項目を実行
harness doctor
```

`harness doctor` は以下を確認する:

| チェック項目 | 内容 |
|------------|------|
| バイナリ | `harness version` が正常に返るか |
| プラグイン設定 | `.claude-plugin/plugin.json` の形式が正しいか |
| hooks 配置 | hooks が正しいパスに存在するか |
| mirror 同期 | skills/ と mirror の内容が一致しているか |
| CLAUDE.md | 必須セクションが存在するか |

問題が検出された場合は修正コマンドを提示する。

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
# インストール確認（Codex CLI は Node.js ベース。Harness 本体とは別物）
which codex || npm install -g @openai/codex

# タイムアウトコマンド確認（macOS）
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")
# macOS の場合: brew install coreutils
```

> **注意**: Harness v4.0 本体（`harness` コマンド）は Node.js 不要の Go バイナリ。
> Codex CLI（`codex` コマンド）は別ツールであり、引き続き Node.js が必要。

**使用パターン**（公式プラグイン経由）:
```bash
bash scripts/codex-companion.sh task --write "タスク内容"
# または stdin 経由
cat /tmp/prompt.md | bash scripts/codex-companion.sh task --write
```

### harness-mem — メモリ設定

Unified Harness Memory の設定を行う。

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

### mirrors — 公開 skill bundle 同期

Windows の `core.symlinks=false` では repository symlink が通常ファイルになり、`harness-*` skill が command 一覧に出なくなることがあります。公開 bundle は実ディレクトリ mirror として同期します。

```bash
./scripts/sync-skill-mirrors.sh
./scripts/sync-skill-mirrors.sh --check
```

更新対象:

- `skills/`
- `codex/.codex/skills/`
- `opencode/skills/`

### agents — エージェント設定

agents/ の3エージェント構成を設定する。

```
agents/
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

## Plugin インストール (v2.1.71+ Marketplace)

v2.1.71 で Marketplace の安定性が大幅に改善された。

### 推奨インストール方式

```bash
# @ref 形式でバージョン固定（推奨）
claude plugin install owner/repo@v4.0.0

# 最新版
claude plugin install owner/repo
```

`owner/repo@vX.X.X` 形式を推奨。`@ref` パーサー修正により、タグ・ブランチ・コミットハッシュいずれも正確に解決される。

### アップデート

```bash
claude plugin update owner/repo
```

v2.1.71 で update 時の merge conflict が修正され、安定したアップデートが可能になった。

### その他の改善点

- MCP server 重複排除: 同一 MCP サーバーの多重登録を自動防止
- `/plugin uninstall` が `settings.local.json` を使用: ユーザーローカル設定に正確に反映

## Maintenance — ファイル整理

定期メンテナンスタスク:

| タスク | コマンド |
|--------|---------|
| 古いログ削除 | `find .claude/logs -mtime +30 -delete` |
| Plans.md 圧縮 | 完了タスクをアーカイブセクションに移動 |
| 古いトレース削除 | `tail -1000 .claude/state/agent-trace.jsonl > /tmp/trace && mv /tmp/trace .claude/state/agent-trace.jsonl` |

## 関連スキル

- `harness-plan` — セットアップ後にプロジェクト計画を作成
- `harness-work` — セットアップ後にタスクを実行
- `harness-review` — セットアップ設定をレビュー
