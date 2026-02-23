# Release Notes Template

GitHub Release Notes の標準テンプレートと作成ガイドライン。

> **フォーマット選択ガイド**
>
> - **最小構成** (`What's Changed` + `Before/After` + `Added/Changed/Fixed` + フッター):
>   [.claude/rules/github-release.md](../../.claude/rules/github-release.md) を参照。シンプルなリリースに適している。
>
> - **拡張構成**（このファイル）: `Highlights` / `Breaking Changes` / `Notable Changes` / `Full Changelog` の4セクション構造。
>   変更量が多いリリースや Breaking Changes を含むリリースに適している。
>
> どちらの構成でも **Before/After テーブル、英語、フッター** は必須。

> **重要**: Release Notes は **英語必須**。フッター省略禁止。Before/After テーブル省略禁止。

---

## 4セクション構造テンプレート

```markdown
## What's Changed

**{one-line value summary — what users gain from this release}**

### Before / After

| Before | After |
|--------|-------|
| {old behavior or limitation} | {new behavior or improvement} |
| {previous state} | {new state} |

---

## Highlights

- **{Feature/Fix name}**: {1-3 sentence description of impact and benefit}
- **{Feature/Fix name}**: {1-3 sentence description}
- **{Feature/Fix name}**: {1-3 sentence description}

## Breaking Changes

> **Migration required** — see steps below before upgrading.

- **{Change name}**: {description of what changed and why}

### Migration Steps

1. {Step 1}
2. {Step 2}
3. {Step 3}

## Notable Changes

### Added

- **{Feature}**: {description}
  - {detail}
  - {detail}

### Changed

- **{Change}**: {description}

### Fixed

- **{Fix}**: {description}

### Removed

- **{Item}**: {what was removed and alternative if any}

### Deprecated

- **{Item}**: {will be removed in vX.Y.Z — use {alternative} instead}

### Security

- **{CVE/Issue}**: {severity and description}

## Full Changelog

**Full Changelog**: https://github.com/{OWNER}/{REPO}/compare/v{PREV}...v{VERSION}

---

Generated with [Claude Code](https://claude.com/claude-code)
```

---

## セクション別ガイドライン

### What's Changed（必須）

- 最初の **太字の1行** はユーザー視点でのバリュー要約
- 技術的変更の列挙ではなく「ユーザーが何を得るか」を述べる
- Before/After テーブルは省略不可（変更がない場合も「No breaking changes」を記載）

**良い例**:
```
**`/work --full` now automates implement → self-review → improve → commit in parallel**
```

**悪い例**:
```
**Added task-worker.md and --full option**
```

### Highlights（推奨 1〜3件）

- このリリースで最も重要な変更を 1〜3件に絞る
- 各 Highlight は 1〜3文で価値を説明する
- Breaking Changes がある場合は必ず Highlights に含める

### Breaking Changes（該当時のみ）

- ユーザーの既存ワークフローに影響する変更は必ず記載
- **移行手順を必ず付ける**（手順なしは禁止）
- 該当なしの場合はセクションごと省略する

### Notable Changes（該当カテゴリのみ）

Keep a Changelog 準拠のカテゴリ:

| カテゴリ | 対象 |
|----------|------|
| `Added` | 新機能 |
| `Changed` | 既存機能の変更 |
| `Fixed` | バグ修正 |
| `Removed` | 廃止・削除された機能 |
| `Deprecated` | 将来削除予定の機能 |
| `Security` | セキュリティ修正 |

- 該当なしのカテゴリは省略する
- 各エントリは **太字の機能名**: 説明の形式

### Full Changelog（必須）

Compare リンクは以下の形式:
```
https://github.com/{OWNER}/{REPO}/compare/v{PREV}...v{VERSION}
```

リポジトリ情報の取得:
```bash
gh repo view --json nameWithOwner -q '.nameWithOwner'
```

---

## Before / After テーブル パターン集

### 機能追加の場合

| Before | After |
|--------|-------|
| Not available | {New feature} supported |
| Manual {task} required | `/{skill}` automates {task} |

### バグ修正の場合

| Before | After |
|--------|-------|
| {Error/unexpected behavior} when {condition} | {Correct behavior} |
| `{command}` fails with `{error}` | `{command}` works correctly |

### パフォーマンス改善の場合

| Before | After |
|--------|-------|
| {operation} takes ~{old time} | {operation} takes ~{new time} ({X}x faster) |

### Breaking Change の場合

| Before | After |
|--------|-------|
| `{old_api}` | `{new_api}` (migration required) |
| Config key `{old_key}` | Config key `{new_key}` |

---

## 禁止事項

- Release Notes 内で日本語を使用すること
- Before/After テーブルを省略すること
- フッター (`Generated with [Claude Code](...)`) を省略すること
- ユーザー視点のバリュー説明なしに技術変更のみを列挙すること
- Breaking Change に移行手順を付けないこと
- Highlights を 3件超にすること（重要度の高いものを厳選）

---

## 参照

- [.claude/rules/github-release.md](../../.claude/rules/github-release.md) — フォーマットルール（正本）
- [references/changelog-format.md](./changelog-format.md) — CHANGELOG.md の形式
- 良い例: v2.8.0, v2.8.2, v2.9.1 の GitHub Releases
