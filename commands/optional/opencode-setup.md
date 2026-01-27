---
description: opencode.ai 用にプロジェクトをセットアップ
description-en: Setup project for opencode.ai compatibility
---

# /opencode-setup - OpenCode セットアップ

現在のプロジェクトに opencode.ai 互換のコマンド、スキル、設定ファイルを生成します。

## VibeCoder Quick Reference

- "**opencode でも使いたい**" → このコマンド
- "**GPT でも Harness 使いたい**" → opencode セットアップ
- "**マルチ LLM 開発したい**" → opencode 互換設定
- "**スキルも opencode で使いたい**" → このコマンドで自動対応

## Deliverables

- `.opencode/commands/` - opencode 用コマンド（Impl + PM）
  - `core/` - コアコマンド（/work, /plan-with-agent 等）
  - `optional/` - オプションコマンド
  - `pm/` - PM コマンド（OpenCode を PM として使う場合）
  - `handoff/` - ハンドオフコマンド
- `.claude/skills/` - opencode 互換スキル（NotebookLM、レビュー等）
- `AGENTS.md` - opencode 用ルールファイル（CLAUDE.md 全文）
---

## Usage

```bash
/opencode-setup
```

---

## Execution Flow

### Step 1: 確認

> 🔧 **opencode.ai 互換ファイルを生成します**
>
> 以下のファイルが作成されます：
> - `.opencode/commands/` - Harness コマンド
> - `.claude/skills/` - Harness スキル（NotebookLM、レビュー等）
> - `AGENTS.md` - ルールファイル（CLAUDE.md 全文）
>
> 続行しますか？ (y/n)

**ユーザーの回答を待つ**

### Step 2: ディレクトリ作成

```bash
mkdir -p .opencode/commands/core
mkdir -p .opencode/commands/optional
mkdir -p .opencode/commands/pm
mkdir -p .opencode/commands/handoff
mkdir -p .claude/skills
```

### Step 3: テンプレをコピー（必須）

**必ず Bash で以下を実行**して、テンプレを直接コピーします。
LLM が内容を自己生成してはいけません。

```bash
bash ./scripts/opencode-setup-local.sh
```

### Step 4: コピー内容の確認

```bash
ls -la .opencode/commands
ls -la .claude/skills
ls -la AGENTS.md
```

### Step 5: 完了メッセージ

> ✅ **OpenCode セットアップ完了**
>
> 📁 **生成されたファイル:**
> - `.opencode/commands/` - Harness コマンド
>   - `core/` - コアコマンド（/work, /plan-with-agent 等）
>   - `optional/` - オプションコマンド
>   - `pm/` - PM コマンド（/start-session, /plan-with-cc 等）
>   - `handoff/` - ハンドオフコマンド
> - `.claude/skills/` - Harness スキル
> - `AGENTS.md` - ルールファイル（CLAUDE.md 全文）
>
> **利用可能なスキル:**
> - `docs` - ドキュメント生成（NotebookLM YAML、スライド）
> - `impl` - 機能実装
> - `review` - コードレビュー
> - `verify` - ビルド検証・エラー復旧
> - `auth` - 認証・決済（Clerk, Stripe）
> - `deploy` - デプロイ（Vercel, Netlify）
>
> **使い方 (Impl モード - Claude Code で実装する場合):**
> ```bash
> # opencode を起動してタスク実行
> opencode
> /work
> ```
>
> **使い方 (PM モード - OpenCode で計画管理する場合):**
> ```bash
> # opencode を起動してセッション開始
> opencode
> /start-session
> /plan-with-cc
> /handoff-to-claude  # Claude Code への依頼生成
> ```
>
> **ドキュメント:** https://github.com/Chachamaru127/claude-code-harness

---

## Notes

- 既存の `.opencode/` ディレクトリがある場合は上書き確認
- `AGENTS.md` が既存の場合はバックアップを作成
- `.claude/skills/` が既存の場合はバックアップを作成
- **Windows ユーザー**: シンボリックリンクは管理者権限が必要なため、コピーを推奨

---

## Related Commands

- `/harness-init` - Harness プロジェクト初期化
- `/harness-update` - Harness 更新
