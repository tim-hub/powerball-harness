# Changelog Format Reference

CHANGELOG.md の書き方と更新手順。[Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/) 準拠。

---

## 基本構造

```markdown
# Changelog

Change history for <project-name>.

> **Writing Guidelines**: Focus on user-facing changes. Keep internal fixes brief.

## [Unreleased]

### 🎯 What's Changed for You

**One-line summary of the phase/milestone.**

| Before | After |
|--------|-------|
| Previous state description | New state description |

### Added
- **Feature name**: Description

### Changed
- **Change**: Description

### Deprecated
- **Feature**: Description

### Removed
- **Feature**: Description

### Fixed
- **Bug**: Description

### Security
- **CVE/Issue**: Description

---

## [X.Y.Z] - YYYY-MM-DD

### 🎯 What's Changed for You

...
```

---

## 6セクション (Keep a Changelog 準拠)

| セクション | 用途 |
|-----------|------|
| `Added` | 新機能の追加 |
| `Changed` | 既存機能の変更 |
| `Deprecated` | 将来削除予定の機能 |
| `Removed` | 削除された機能 |
| `Fixed` | バグ修正 |
| `Security` | 脆弱性対応 |

使われないセクションは省略可。

---

## Harness 固有: 🎯 What's Changed for You

各バージョンに **必ず** `### 🎯 What's Changed for You` セクションを設ける。

### 形式

```markdown
### 🎯 What's Changed for You

**Phase N: 一行でフェーズ・マイルストーンの価値を説明する。**

| Before | After |
|--------|-------|
| 変更前の状態（ユーザー視点） | 変更後の状態（ユーザー視点） |
| ... | ... |
```

### ルール

- **太字サマリ** は必須（1行）
- Before/After テーブルは必須（最低1行）
- ユーザー視点で書く（実装詳細ではなく効果を書く）
- 技術的な補足は `Added/Changed/Fixed` セクションに書く

---

## デュアルフォーマット (CHANGELOG.md + CHANGELOG_ja.md)

このプロジェクトは英語と日本語の2ファイル体制。

| ファイル | 言語 | 内容 |
|---------|------|------|
| `CHANGELOG.md` | 英語 | 主ファイル。Keep a Changelog 準拠 |
| `CHANGELOG_ja.md` | 日本語 | 日本語版。同じ構造で日本語表記 |

### 更新手順

1. `CHANGELOG.md` を更新（英語）
2. `CHANGELOG_ja.md` を同内容で更新（日本語訳）
3. `[Unreleased]` の内容を新バージョン番号に移動

---

## ISO 8601 日付フォーマット

```
YYYY-MM-DD
例: 2026-02-23
```

---

## [Unreleased] セクション

- リリース前の変更をここに蓄積する
- リリース時に `## [X.Y.Z] - YYYY-MM-DD` として切り出す
- リリース後の `[Unreleased]` は空でよい（見出しだけ残す）

---

## Compare リンク

ファイル末尾に compare リンクを追加する（オプション）:

```markdown
[Unreleased]: https://github.com/OWNER/REPO/compare/vX.Y.Z...HEAD
[X.Y.Z]: https://github.com/OWNER/REPO/compare/vA.B.C...vX.Y.Z
[A.B.C]: https://github.com/OWNER/REPO/releases/tag/vA.B.C
```

---

## 更新チェックリスト

リリース時に確認:

- [ ] `[Unreleased]` の内容がすべて移動済み
- [ ] バージョン番号が正しい (`X.Y.Z`)
- [ ] 日付が ISO 8601 フォーマット (`YYYY-MM-DD`)
- [ ] `🎯 What's Changed for You` セクションあり
- [ ] Before/After テーブルあり
- [ ] `CHANGELOG_ja.md` も同内容で更新済み
- [ ] Compare リンクが更新済み（存在する場合）
