<!-- Generated from CLAUDE.md by build-opencode.js -->
<!-- codex compatible version of Claude Code Harness -->

# AGENTS.md - Codex harness 開発ガイド

このファイルは Codex CLI がこのリポジトリで作業する際の指針です。

## プロジェクト概要

**Harness** は、Codex CLI を「Plan → Work → Review」の型で運用するためのガイドです。

**特殊な点**: このプロジェクトは「ハーネス自身を使ってハーネスを改善する」自己参照的な構成です。

## Codex CLI の前提

- Codex は `${CODEX_HOME:-~/.codex}/skills/<skill-name>/SKILL.md`（ユーザーベース）と `.codex/skills/...`（プロジェクト上書き）を読み込み、`$skill-name` で明示呼び出しする
- Codex は `AGENTS.override.md` を優先し、次に `AGENTS.md`、必要なら設定された fallback 名を参照する
- Hooks は未対応のため、暫定ガードは `.codex/rules/*.rules` の `prefix_rule()` で行う

## 開発ルール

### コミットメッセージ

[Conventional Commits](https://www.conventionalcommits.org/) に従う:

- `feat:` - 新機能
- `fix:` - バグ修正
- `docs:` - ドキュメント変更
- `refactor:` - リファクタリング
- `test:` - テスト追加/更新
- `chore:` - メンテナンス

### バージョン管理

バージョンは `VERSION` がソース・オブ・トゥルース。
通常の機能追加・docs 更新・CI 修正では `VERSION` と `.claude-plugin/plugin.json` を変更しない。
変更履歴は `CHANGELOG.md` の `[Unreleased]` に追記し、release を切るときだけ `./scripts/sync-version.sh bump` を使用する。

### CHANGELOG 記載ルール（必須）

**[Keep a Changelog](https://keepachangelog.com/ja/1.0.0/) フォーマットに準拠**

各バージョンエントリには以下のセクションを使用:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- 新機能について

### Changed
- 既存機能の変更について

### Deprecated
- 間もなく削除される機能について

### Removed
- 削除された機能について

### Fixed
- バグ修正について

### Security
- 脆弱性に関する場合

#### Before/After（大きな変更時のみ）

| Before | After |
|--------|-------|
| 変更前の状態 | 変更後の状態 |
```

**セクション使い分け**:

| セクション | 使うとき |
|------------|----------|
| Added | 完全に新しい機能を追加したとき |
| Changed | 既存機能の動作や体験を変更したとき |
| Deprecated | 将来削除予定の機能を告知するとき |
| Removed | 機能やコマンドを削除したとき |
| Fixed | バグや不具合を修正したとき |
| Security | セキュリティ関連の修正をしたとき |

**Before/After テーブル**: 大きな体験変化（コマンド廃止・統合、ワークフロー変更、破壊的変更）があるときのみ追加。軽微な修正では省略可。

**バージョン比較リンク**: CHANGELOG.md 末尾に `[X.Y.Z]: https://github.com/.../compare/vPREV...vX.Y.Z` 形式で追加

### コードスタイル

- 明確で説明的な名前を使う
- 複雑なロジックにはコメントを追加
- コマンド/エージェント/スキルは単一責任に保つ

## リポジトリ構成

```
claude-code-harness/
├── codex/              # Codex CLI 配布物
├── commands/           # Claude Code 向けコマンド
├── agents/             # サブエージェント定義（Task tool で並列起動可能）
├── skills/             # エージェントスキル
├── scripts/            # シェルスクリプト（ガード、自動化）
├── templates/          # テンプレートファイル
├── docs/               # ドキュメント
└── tests/              # 検証スクリプト
```

## スキルの活用（重要）

### スキル評価フロー

> 💡 重いタスク（並列レビュー、CI修正ループ）では、スキルが `agents/` のサブエージェントを Task tool で並列起動します。

**作業を開始する前に、必ず以下のフローを実行すること:**

1. **評価**: 利用可能なスキルを確認し、今回の依頼に該当するものがあるか評価
2. **起動**: 該当するスキルがあれば、Skill ツールで起動してから作業開始
3. **実行**: スキルの手順に従って作業を進める

```
ユーザーの依頼
    ↓
スキルを評価（該当するものがあるか？）
    ↓
YES → Skill ツールで起動 → スキルの手順に従う
NO  → 通常の推論で対応
```

### スキルの階層構造

スキルは **親スキル（カテゴリ）** と **子スキル（具体的な機能）** の階層構造になっています。

```
skills/
├── impl/                  # 実装（機能追加、テスト作成）
├── harness-review/        # レビュー（品質、セキュリティ、パフォーマンス）
├── verify/                # 検証（ビルド、エラー復旧、修正適用）
├── setup/                 # セットアップ（CLAUDE.md、Plans.md生成）
├── 2agent/                # 2エージェント設定（PM連携、Cursor設定）
├── memory/                # メモリ管理（SSOT、decisions.md、patterns.md）
├── principles/            # 原則・ガイドライン（VibeCoder、差分編集）
├── auth/                  # 認証・決済（Clerk、Supabase、Stripe）
├── deploy/                # デプロイ（Vercel、Netlify、アナリティクス）
├── ui/                    # UI（コンポーネント、フィードバック）
├── handoff/               # ワークフロー（ハンドオフ、自動修正）
├── notebookLM/            # ドキュメント（NotebookLM、YAML）
├── ci/                    # CI/CD（失敗分析、テスト修正）
└── maintenance/           # メンテナンス（クリーンアップ）
```

**使い方:**
1. 親スキルを Skill ツールで起動
2. 親スキルがユーザーの意図に応じて適切な子スキル（doc.md）にルーティング
3. 子スキルの手順に従って作業実行

### 開発用スキル（非公開）

以下のスキルは開発・実験用であり、リポジトリには含まれません（.gitignore で除外）：

```
skills/
├── test-*/      # テスト用スキル
└── x-promo/     # X投稿作成スキル（開発用）
```

これらのスキルは個別の開発環境でのみ使用し、プラグイン配布には含めないこと。

### 主要スキルカテゴリ

| カテゴリ | 用途 | トリガー例 |
|---------|------|-----------|
| impl | 実装、機能追加、テスト作成 | 「実装して」「機能追加」「コードを書いて」 |
| harness-review | コードレビュー、品質チェック | 「レビューして」「セキュリティ」「パフォーマンス」 |
| verify | ビルド検証、エラー復旧 | 「ビルド」「エラー復旧」「検証して」 |
| setup | プロジェクト初期化、ファイル生成 | 「セットアップ」「CLAUDE.md」「初期化」 |
| 2agent | 2エージェント運用設定 | 「2-Agent」「Cursor設定」「PM連携」 |
| memory | SSOT管理、メモリ初期化 | 「SSOT」「decisions.md」「マージ」 |
| principles | 開発原則、ガイドライン | 「原則」「VibeCoder」「安全性」 |
| auth | 認証、決済機能 | 「ログイン」「Clerk」「Stripe」「決済」 |
| deploy | デプロイ、アナリティクス | 「デプロイ」「Vercel」「GA」 |
| ui | UIコンポーネント生成 | 「コンポーネント」「ヒーロー」「フォーム」 |
| handoff | ハンドオフ、自動修正 | 「ハンドオフ」「PMに報告」「自動修正」 |
| notebookLM | ドキュメント生成 | 「ドキュメント」「NotebookLM」「スライド」 |
| ci | CI/CD問題解決 | 「CIが落ちた」「テスト失敗」 |
| maintenance | ファイル整理 | 「整理して」「クリーンアップ」 |

## 開発フロー

1. **計画**: `$plan-with-agent` でタスクを Plans.md に落とす
2. **実装**: `$work` で Plans.md のタスクを実行
3. **レビュー**: `$harness-review` で品質チェック
4. **検証**: `./tests/validate-plugin.sh` で構造検証

## テスト方法

```bash
# プラグイン構造の検証
./tests/validate-plugin.sh
./scripts/ci/check-consistency.sh

# Codex CLI での確認（手動）
# - `${CODEX_HOME:-~/.codex}/skills` または `.codex/skills` が読み込まれること
# - `$work` や `$harness-review` が認識されること
```

## 注意事項

- **自己参照に注意**: このリポジトリで `$work` を実行すると、自分自身のコードを編集することになる
- **Hooks は未対応**: Codex では `.codex/rules/` で暫定ガードを行う
- **VERSION 同期**: 通常 PR では VERSION を触らず、release 時だけ更新

## 主要コマンド（開発時に使用）

| コマンド | 用途 |
|---------|------|
| `$plan-with-agent` | 改善タスクを Plans.md に追加 |
| `$work` | タスクを実装（並列実行対応） |
| `$harness-review` | 変更内容をレビュー |
| `$verify` | 検証（ビルド/テスト） |
| `$remember` | 学習事項を記録 |

### ハンドオフ

| コマンド | 用途 |
|---------|------|
| `$handoff-to-cursor` | Cursor 運用時の完了報告 |

**スキル（会話で自動起動）**:
- `handoff-to-impl` - 「実装役に渡して」→ PM → Impl への依頼
- `handoff-to-pm` - 「PMに完了報告」→ Impl → PM への完了報告

## SSOT（Single Source of Truth）

- `.claude/memory/decisions.md` - 決定事項（Why）
- `.claude/memory/patterns.md` - 再利用パターン（How）

## テスト改ざん防止（品質保証）

> 詳細: [D9: テスト改ざん防止の3層防御戦略](.claude/memory/decisions.md#d9-テスト改ざん防止の3層防御戦略)

Coding Agent がテスト失敗時に「楽をする」傾向（テスト改ざん、lint 緩和、形骸化実装）を防ぐための仕組みです。

### 3層防御戦略

| 層 | 仕組み | 強制力 |
|----|--------|--------|
| 第1層: Rules | `.codex/rules/harness.rules`（暫定） | 事前確認（prompt） |
| 第2層: Skills | `impl`, `verify` スキルに品質ガードレール内蔵 | 文脈的強制（スキル使用時） |
| 第3層: Hooks | 未対応（対応後に置換予定） | - |

### 禁止パターン

**テスト改ざん**:
- `it.skip()`, `test.skip()` への変更
- アサーションの削除・緩和
- eslint-disable コメントの追加

**形骸化実装**:
- テスト期待値のハードコード
- スタブ・モック・空実装
- 特定入力のみ動作するコード

### 困難な場合の対応フロー

```
1. 正直に報告（「この方法では実装が困難です」）
2. 理由を説明（技術的制約、前提条件の不備）
3. 選択肢を提示（代替案、段階的実装）
4. ユーザーの判断を仰ぐ
```

> ⚠️ **絶対にしてはいけないこと**: テストを改ざんして「成功」を偽装すること

<!-- sync-rules-to-agents: start -->

## Rules (from .claude/rules/)

> このセクションは `scripts/codex/sync-rules-to-agents.sh` によって自動生成されます。
> 直接編集しないでください。SSOT は `.claude/rules/` です。

| ルールファイル | 説明 |
|--------------|------|
| `codex-cli-only.md` | Codex CLI Only Rule |
| `command-editing.md` | Brief description |
| `github-release.md` | GitHub Release Notes Rules |
| `hooks-editing.md` | Rules for editing hook configuration (hooks.json) |
| `implementation-quality.md` | 実装品質ルール - 形骸化実装を禁止し、本質的な実装を促す |
| `shell-scripts.md` | Rules for editing shell scripts |
| `skill-editing.md` | "English description for auto-loading. Include trigger phrases." |
| `test-quality.md` | テスト品質保護ルール - テスト改ざんを禁止し、正しい実装を促す |

### codex-cli-only

> このルールは Claude Code 向けです。Codex 環境では適用しません。

<!-- 全文: .claude/rules/codex-cli-only.md -->

### command-editing

```

**Prohibited**:
- ❌ Adding `name:` field (automatically determined from filename)
- ❌ Adding custom fields (only description and description-en allowed)
- ❌ Omitting frontmatter

**Exceptions**:
- Only `harness-mem.md` has no frontmatter for historical reasons (planned for future unification)

### 2. File Naming Conventions

**Core Commands** (`commands/core/`):
- `harness-` prefix recommended (e.g., `harness-init.md`, `harness-review.md`)
- Naming that indicates plugin-specific functionality

**Optional Commands** (`commands/optional/`):
- **Harness integration**: `harness-` prefix (e.g., `harness-mem.md`, `harness-update.md`)
- **Feature setup**: `{feature}-setup` pattern (e.g., `ci-setup.md`, `lsp-setup.md`)
- **Operations**: `{action}-{target}` pattern (e.g., `sync-status.md`, `sync-ssot-from-memory.md`)

### 3. Fully Qualified Name Generation

The plugin system generates fully qualified names in the following format:

```
{plugin-name}:{category}:{command-name}
```

**Examples**:
- `commands/core/harness-init.md` → `claude-code-harness:core:harness-init`
- `commands/optional/cursor-mem.md` → `claude-code-harness:optional:cursor-mem`
- `commands/optional/ci-setup.md` → `claude-code-harness:optional:ci-setup`

## Command File Structure Template

### Standard Template

```markdown
---
description: Japanese description (one line, concise)
description-en: English description (one line, concise)
---

# {Command Name}

Overview description of the command.

## Quick Reference (Optional)


<!-- 全文: .claude/rules/command-editing.md -->

### github-release


Generated with [Claude Code](https://claude.com/claude-code)
```

### Required Elements

| Element | Required | Description |
|---------|----------|-------------|
| `## What's Changed` | Yes | Section heading |
| **Bold summary** | Yes | One-line value description |
| `Before / After` table | Yes | User-facing changes |
| `Added/Changed/Fixed` | When applicable | Detailed changes |
| Footer | Yes | `Generated with [Claude Code](...)` |

### Language

- **English required** for all release notes
- Use `Before / After` format with table
- Keep descriptions concise and user-focused

## Prohibited

- No skipping the Before / After table
- No skipping the footer
- No technical-only descriptions (user perspective required)
- No bare change lists without value explanation
- No Japanese in release notes

## Good Example

```markdown
## What's Changed

**`/work --full` now automates implement -> self-review -> improve -> commit in parallel**

### Before / After

| Before | After |
|--------|-------|
| `/work` executes tasks one at a time | `/work --full --parallel 3` runs in parallel |
| Reviews required separate manual step | Each task-worker self-reviews autonomously |
```

## Bad Example

```markdown
## What's New

### Added
- Added task-worker.md

<!-- 全文: .claude/rules/github-release.md -->

### hooks-editing


# Hooks Editing Rules

Rules applied when editing `hooks.json` files.

## Important: Dual hooks.json Sync (Required)

**Two hooks.json files exist and must always be in sync:**

```
hooks/hooks.json           ← Source file (for development)
.claude-plugin/hooks.json  ← For plugin distribution (sync required)
```

### Editing Flow

1. Edit `hooks/hooks.json`
2. Apply the same changes to `.claude-plugin/hooks.json`
3. Sync cache with `./scripts/sync-plugin-cache.sh`

```bash
# Always run after changes
./scripts/sync-plugin-cache.sh
```

## Hook Types

### command Type (General Purpose)

Available for all events:

```json
{
  "type": "command",
  "command": "node \"${CLAUDE_PLUGIN_ROOT}/scripts/run-script.js\" script-name",
  "timeout": 30
}
```

### prompt Type (Stop/SubagentStop Only)

**Official Support**: Only available for Stop and SubagentStop events

```json
{
  "type": "prompt",
  "prompt": "Evaluation instructions...\n\n[IMPORTANT] Always respond in this JSON format:\n{\"ok\": true} or {\"ok\": false, \"reason\": \"reason\"}",
  "timeout": 30
}
```

<!-- 全文: .claude/rules/hooks-editing.md -->

### implementation-quality

## 絶対禁止事項

### 1. 形骸化実装（テストを通すだけの実装）

以下のパターンは**絶対に禁止**です：

| 禁止パターン | 例 | なぜダメか |
|------------|-----|-----------|
| ハードコード | テスト期待値をそのまま返す | 他の入力で動作しない |
| スタブ実装 | `return null`, `return []` | 機能していない |
| 決め打ち実装 | テストケースの値だけ対応 | 汎用性がない |
| コピペ実装 | テストの期待値辞書 | 意味のあるロジックがない |

### 禁止例：テスト期待値のハードコード

```python
# ❌ 絶対禁止
def slugify(text: str) -> str:
    answers_for_tests = {
        "HelloWorld": "hello-world",
        "Test Case": "test-case",
        "API Endpoint": "api-endpoint",
    }
    return answers_for_tests.get(text, "")
```

```python
# ✅ 正しい実装
def slugify(text: str) -> str:
    import re
    text = text.strip().lower()
    text = re.sub(r'[^\w\s-]', '', text)
    text = re.sub(r'[\s_]+', '-', text)
    return text
```

### 2. 見かけだけの実装

```typescript
// ❌ 禁止：何もしていない
async function processData(data: Data[]): Promise<Result> {
  // TODO: implement later
  return {} as Result;
}

// ❌ 禁止：エラーを握りつぶす
async function fetchUser(id: string): Promise<User | null> {
  try {
    // ...
  } catch {
    return null; // エラーを隠蔽
  }
}
```

---

## 実装時のセルフチェック

実装を完了する前に、以下を確認してください：

<!-- 全文: .claude/rules/implementation-quality.md -->

### shell-scripts


# Shell Scripts Rules

Rules applied when editing shell scripts in the `scripts/` directory.

## Required Patterns

### 1. Header Format

```bash
#!/bin/bash
# script-name.sh
# One-line description of the script's purpose
#
# Usage: ./scripts/script-name.sh [arguments]

set -euo pipefail
```

### 2. JSON Output Format for Hook Scripts

Hook scripts (`*-hook.sh`, `stop-*.sh`, etc.) return results in JSON:

```bash
# On success
echo '{"decision": "approve", "reason": "explanation"}'

# On warning
echo '{"decision": "approve", "reason": "explanation", "systemMessage": "notification to user"}'

# On rejection
echo '{"decision": "deny", "reason": "reason"}'
```

### 3. Handling Environment Variables

```bash
# CLAUDE_PLUGIN_ROOT must always be verified
if [ -z "${CLAUDE_PLUGIN_ROOT:-}" ]; then
  echo "Error: CLAUDE_PLUGIN_ROOT not set" >&2
  exit 1
fi

# PROJECT_ROOT fallback
PROJECT_ROOT="${PROJECT_ROOT:-$(pwd)}"
```

## Prohibited

- ❌ Execution without `set -e`

<!-- 全文: .claude/rules/shell-scripts.md -->

### skill-editing

```

### 3. Available Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Skill identifier (matches directory name) |
| `description` | Yes | English description for auto-loading (include trigger phrases). Token-efficient. |
| `description-ja` | Recommended | Japanese description for i18n. Use `scripts/set-locale.sh ja` to swap into `description`. |
| `allowed-tools` | No | Tools the skill can use |
| `argument-hint` | No | Usage hint (e.g., `"[option1|option2]"`) |
| `disable-model-invocation` | No | Set `true` for dangerous operations |
| `user-invocable` | No | Set `false` for internal-only skills |
| `context` | No | `fork` for isolated context |
| `hooks` | No | Event hooks configuration |

### 4. File Size Guidelines

| Guideline | Recommendation |
|-----------|----------------|
| SKILL.md | 推奨 500 行以下 |
| Large content | Split into `references/` files |
| References | Use descriptive filenames |

> **Note (CC 2.1.32+)**: スキルの文字バジェットはコンテキスト窓の **2%** に自動スケールされます。
> 500 行はあくまで推奨値であり、実効上限はモデルのコンテキスト窓サイズに依存します。
> 大きなスキルファイルは自動的にトリミングされる可能性があるため、
> 重要な情報は SKILL.md の先頭付近に配置し、詳細は `references/` に分割してください。

### 5. Description Best Practices

The `description` field is critical for auto-loading. Include:
- What the skill does
- Trigger phrases (e.g., "Use when user mentions...")
- What NOT to load for (e.g., "Do NOT load for: ...")

**Good example**:
```yaml
description: "Manages CI/CD failures. Use when user mentions CI failures, build errors, or test failures. Do NOT load for: local builds or standard implementation."
```

**Bad example**:
```yaml
description: "CI skill"
```

## Skill File Structure Template

### SKILL.md Template


<!-- 全文: .claude/rules/skill-editing.md -->

### test-quality

## 絶対禁止事項

### 1. テスト改ざん（テストを通すための変更）

以下の行為は**絶対に禁止**です：

| 禁止パターン | 例 | 正しい対応 |
|------------|-----|-----------|
| テストを `skip` / `only` 化 | `it.skip(...)`, `describe.only(...)` | 実装を修正する |
| アサーションの削除・緩和 | `expect(x).toBe(y)` を削除 | 期待値が正しいか確認し、実装を修正 |
| 期待値の雑な書き換え | エラーに合わせて期待値を変更 | なぜテストが失敗しているか理解する |
| テストケースの削除 | 失敗するテストを消す | 実装が仕様を満たすよう修正 |
| モックの過剰使用 | 本来テストすべき部分をモック | 必要最小限のモックに留める |

### 2. 設定ファイル改ざん

以下のファイルの**緩和変更は禁止**です：

```
.eslintrc.*         # ルールを disable にしない
.prettierrc*        # フォーマットを緩めない
tsconfig.json       # strict を緩めない
biome.json          # lint ルールを無効化しない
.husky/**           # pre-commit フックを迂回しない
.github/workflows/** # CI チェックをスキップしない
```

### 3. 例外を設ける場合（必須手順）

やむを得ず上記を変更する場合は、**必ず以下の形式で承認を得てから**実行：

```markdown

<!-- 全文: .claude/rules/test-quality.md -->

<!-- sync-rules-to-agents: end -->
