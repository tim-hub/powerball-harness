---
name: harness-release
description: "Harness v3 統合リリーススキル。CHANGELOG・バージョンバンプ・タグ・GitHub Release・mirror同期・検証を自動化。以下で起動: リリース、バージョンバンプ、タグ作成、公開、/harness-release。実装・コードレビュー・プランニング・セットアップには使わない。"
description-en: "Unified release skill for Harness v3. CHANGELOG, version bump, tag, GitHub Release, mirror sync, and validation automation. Use when user mentions: release, version bump, create tag, publish, /harness-release. Do NOT load for: implementation, code review, planning, or setup."
description-ja: "Harness v3 統合リリーススキル。CHANGELOG・バージョンバンプ・タグ・GitHub Release・mirror同期・検証を自動化。以下で起動: リリース、バージョンバンプ、タグ作成、公開、/harness-release。実装・コードレビュー・プランニング・セットアップには使わない。"
allowed-tools: ["Read", "Write", "Edit", "Bash"]
argument-hint: "[patch|minor|major|--dry-run|--announce|--complete]"
context: fork
effort: high
---

# Harness Release (v3)

Harness v3 の統合リリーススキル。
以下の旧スキルを統合:

- `release-har` -- 汎用リリース自動化
- `x-release-harness` -- Harness 専用リリース自動化
- `handoff` -- PM へのハンドオフ・完了報告

## Quick Reference

```bash
/release          # インタラクティブ（バージョン種別を確認）
/release patch    # パッチバージョンバンプ（バグ修正）
/release minor    # マイナーバージョンバンプ（新機能）
/release major    # メジャーバージョンバンプ（破壊的変更）
/release --dry-run   # プレビューのみ（実行しない）
/release --announce  # X (Twitter) 告知も実行
/release --complete  # リリース完了マーキング（タグ後の仕上げ）
```

## Release-only policy

- 通常 PR: `VERSION` / `.claude-plugin/plugin.json` / versioned `CHANGELOG.md` entry は触らない
- 通常 PR の変更履歴: `CHANGELOG.md` の `[Unreleased]` に追記する
- `/release` 実行時だけ version bump、versioned CHANGELOG entry、tag / GitHub Release をまとめて更新する
- `/release --dry-run` でも本番実行と同じ preflight を通し、公開前の危険信号を先に止める

## ブランチポリシー

- **単独開発**: main への直接 push を許容（CI が品質ゲートとして機能）
- **共同開発**: PR 経由のマージが必須
- force push（`--force` / `--force-with-lease`）は常に禁止

## バージョン判定基準（SemVer）

`.claude/rules/versioning.md` に基づく判定フローチャート:

```
既存の動作が壊れる？
├─ Yes → major
└─ No → ユーザーが新しいことをできるようになる？
    ├─ Yes → minor
    └─ No → patch
```

| 変更の種類 | バージョン | 例 |
|-----------|----------|-----|
| スキル定義の文言修正・追記 | **patch** | テンプレート微修正 |
| hooks/scripts のバグ修正 | **patch** | エスケープ修正 |
| 新スキル/フラグ/エージェント追加 | **minor** | `--dual`、新スキル |
| CC 新バージョン互換対応 | **minor** | CC v2.1.90 対応 |
| 破壊的変更（旧スキル廃止、フォーマット非互換） | **major** | Plans.md v1 削除 |

**バッチリリースの推奨**: 同日に複数変更がある場合は 1 つの minor にまとめる。同日 2 回以上の minor バンプは禁止。

## NPM 配布について

このプロジェクトは Claude Code プラグインであり、npm パッケージとしては配布しない。
ルートに `package.json` は存在しない（`core/package.json` は内部 TypeScript ビルド用）。
バージョン管理の対象は以下の 2 ファイルのみ:

- `VERSION` -- 正本
- `.claude-plugin/plugin.json` -- プラグインマニフェスト

## 配布面と Mirror 同期

`skills/` が SSOT（Single Source of Truth）。以下の 2 配布面が mirror として同期される:

| 配布面 | パス | 対象ユーザー |
|--------|------|------------|
| Claude | `skills/harness-release/` | Claude Code ユーザー |
| Codex | `codex/.codex/skills/harness-release/` | Codex CLI ユーザー |
| OpenCode | `opencode/skills/harness-release/` | OpenCode ユーザー |

**重要**: `skills/` を編集したら、リリース前に必ず mirror を同期する:

```bash
./scripts/sync-skill-mirrors.sh
```

検証のみ（書き換えなし）:

```bash
./scripts/sync-skill-mirrors.sh --check
```

## 日本語対応（i18n）

スキルの description フィールドを日英切替できる。リリース前にロケール設定が意図通りか確認する:

```bash
# 日本語に設定（description-ja → description）
./scripts/i18n/set-locale.sh ja

# 英語に設定（description-en → description）
./scripts/i18n/set-locale.sh en
```

現在のデフォルト: description は日本語（`description-ja` と同一）。`description-en` は英語バックアップとして常に保持する。

## 実行フロー

### Phase 0: Pre-flight チェック（必須）

```bash
# 1. 必須ツール確認
command -v gh &>/dev/null || echo "gh なし: GitHub Release はスキップ"
command -v jq &>/dev/null || echo "jq なし: plugin.json 更新に必要"

# 2. vendor-neutral preflight（本実行 / dry-run 共通）
bash scripts/release-preflight.sh

# 3. プラグイン構造検証
bash tests/validate-plugin.sh

# 4. 整合性チェック
bash scripts/ci/check-consistency.sh

# 5. mirror 同期状態の確認
bash scripts/sync-skill-mirrors.sh --check
```

`scripts/release-preflight.sh` は以下を検証する:

- working tree が clean か
- `CHANGELOG.md` に `[Unreleased]` があるか
- `.env.example` と `.env` の差分（managed secrets 前提では warning 止まり）
- `healthcheck` / `preflight` コマンド（あれば実行）
- `agents/` / `core/` / `hooks/` / `scripts/` の shipped surface に debug / mock / placeholder 残骸がないか
- CI 状態（取得可能な場合）

環境変数でリポジトリごとに調整可能:

- `HARNESS_RELEASE_PROJECT_ROOT`
- `HARNESS_RELEASE_HEALTHCHECK_CMD`
- `HARNESS_RELEASE_CI_STATUS_CMD`

詳細: [docs/release-preflight.md](${CLAUDE_SKILL_DIR}/../../docs/release-preflight.md)

### Phase 1: 現在バージョン取得

```bash
CURRENT=$(cat VERSION 2>/dev/null)
echo "現在のバージョン: $CURRENT"
```

### Phase 2: 新バージョン算出

`scripts/sync-version.sh` は patch bump のみ対応。minor / major は手動で VERSION を書き換える:

```bash
# patch バンプ（x.y.Z → x.y.(Z+1)）
./scripts/sync-version.sh bump

# minor バンプ（手動: x.Y.z → x.(Y+1).0）
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
MINOR=$(echo "$CURRENT" | cut -d. -f2)
NEW_VERSION="$MAJOR.$((MINOR + 1)).0"
echo "$NEW_VERSION" > VERSION
./scripts/sync-version.sh sync

# major バンプ（手動: X.y.z → (X+1).0.0）
CURRENT=$(cat VERSION)
MAJOR=$(echo "$CURRENT" | cut -d. -f1)
NEW_VERSION="$((MAJOR + 1)).0.0"
echo "$NEW_VERSION" > VERSION
./scripts/sync-version.sh sync
```

`sync-version.sh sync` は `VERSION` の値を `.claude-plugin/plugin.json` に反映する。

### Phase 3: CHANGELOG 更新

release entry は、通常 PR で溜めた `[Unreleased]` の変更を versioned section へ確定する。

**詳細 Before/After フォーマット**（日本語）で記述する。
各機能を番号付きセクションに分け、「今まで」と「今後」を具体例付きで説明する。

```markdown
## [X.Y.Z] - YYYY-MM-DD

### テーマ: [変更全体を一言で]

**[ユーザーにとっての価値を1〜2文で]**

---

#### 1. [機能名]

**今まで**: [旧動作を具体的に。ユーザーが「あるある」と感じる課題描写]

**今後**: [新動作を具体的に。何が解決するか]

```
[実際の出力例やコマンド例]
```

#### 2. [次の機能名]

**今まで**: ...

**今後**: ...
```

**CC バージョン統合時のパターン**: 通常の「今まで / 今後」ではなく「CC のアプデ → Harness での活用」形式を使う。
詳細は `.claude/rules/github-release.md` の「CC バージョン統合時の CHANGELOG パターン」を参照。

**書き方のルール**:

| ルール | 説明 |
|--------|------|
| 言語 | **日本語** |
| 各機能を独立セクションに | `#### N. 機能名` で番号付き |
| 「今まで」は課題描写 | ユーザーが体験していた不便を具体的に書く |
| 「今後」は解決を示す | 何がどう変わるか + 具体例（コード/出力） |
| 具体例を必ず含める | コマンド例、出力例、Plans.md のスニペット等 |
| テクニカル詳細は最小限に | ファイル名やステップ番号は「今後」の補足として |
| 長くてOK | 各機能3〜10行。読みやすさが最優先 |

`[Unreleased]` セクションは空にせず、次のリリースに向けて残す:

```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD
...
```

### Phase 4: バージョンファイル更新

```bash
# VERSION は Phase 2 で更新済み
# plugin.json を同期
./scripts/sync-version.sh sync

# 同期確認
./scripts/sync-version.sh check
```

### Phase 5: Mirror 同期

```bash
# skills → codex, opencode への mirror 同期
./scripts/sync-skill-mirrors.sh

# 同期確認
./scripts/sync-skill-mirrors.sh --check
```

### Phase 6: コミット & タグ

```bash
NEW_VERSION=$(cat VERSION)

# ステージング（対象ファイルを明示的に指定）
git add VERSION .claude-plugin/plugin.json CHANGELOG.md
git add skills/ codex/.codex/skills/ opencode/skills/

git commit -m "chore: release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
```

### Phase 7: プッシュ

```bash
git push origin main --tags
```

**注意**: `.github/workflows/release.yml` がタグプッシュを検知し、GitHub Release が未作成の場合は CHANGELOG から自動生成するセーフティネットが動く。手動で GitHub Release を先に作成すれば、ワークフローは自動スキップする。

### Phase 8: GitHub Release 作成

```bash
NEW_VERSION=$(cat VERSION)

gh release create "v$NEW_VERSION" \
  --title "v$NEW_VERSION - タイトル" \
  --notes "$(cat <<'EOF'
## What's Changed

**[変更の概要（英語）]**

### Before / After

| Before | After |
|--------|-------|
| Previous state | New state |

---

## Added

- **Feature**: Description

## Changed

- **Change**: Description

## Fixed

- **Fix**: Description

---

Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

GitHub Release Notes のルール:
- 言語: **英語**（公開リポジトリのため）
- 必須: `## What's Changed`、太字サマリー、Before / After テーブル、フッター
- 詳細フォーマット: `.claude/rules/github-release.md` を参照

リリースノートの検証:

```bash
./scripts/validate-release-notes.sh "v$NEW_VERSION"
```

### Phase 9: リリース完了マーキング

```bash
git commit --allow-empty -m "chore: mark v$NEW_VERSION release complete"
git push origin main
```

この空コミットは「リリース作業が全て完了した」ことを明示するマーカー。

### Phase 10: 告知（`--announce` 指定時のみ）

`/x-announce` スキルを呼び出して X (Twitter) への告知スレッドを生成する:

```
Skill: x-announce
Args: v$NEW_VERSION
```

投稿テキスト 5 本 + Gemini 画像 5 枚を 1 発出力する。

## `--dry-run` モード

`--dry-run` は以下を実行し、実際の変更はしない:

1. Pre-flight チェック（Phase 0）を**全て実行**する
2. バージョン算出を表示する（書き込まない）
3. CHANGELOG のドラフトを表示する（書き込まない）
4. GitHub Release Notes のドラフトを表示する（作成しない）
5. mirror 同期の差分を表示する（書き込まない）

スキップされるもの: VERSION/plugin.json 書き換え、git commit/tag/push、GitHub Release 作成、告知

## `--complete` モード

タグ作成後に「リリース完了」のマーキングだけを行う:

```bash
/release --complete
```

Phase 9 のみを実行する。GitHub Release の作成漏れがないか確認したうえで、完了コミットを打つ。

## デグレチェックリスト

リリース前に以下のデグレを確認する:

| チェック項目 | 確認方法 | 備考 |
|------------|---------|------|
| プラグイン構造 | `tests/validate-plugin.sh` | plugin.json、スキル、フック、スクリプトの検証 |
| 整合性 | `scripts/ci/check-consistency.sh` | テンプレート、バージョン、mirror、CHANGELOG |
| Mirror 同期 | `scripts/sync-skill-mirrors.sh --check` | skills と 2 配布面の一致 |
| Preflight | `scripts/release-preflight.sh` | working tree、CHANGELOG、CI、残骸 |
| リリースノート | `scripts/validate-release-notes.sh vX.Y.Z` | GitHub Release のフォーマット検証 |
| VERSION 同期 | `scripts/sync-version.sh check` | VERSION と plugin.json の一致 |
| ガードレール | `core/src/guardrails/rules.ts` の R01-R13 | TypeScript ルールの健全性 |
| タグ連続性 | `git tag --sort=-version:refname \| head -5` | 欠番がないこと |
| ロケール | description と description-ja の一致 | `set-locale.sh` で切替可能 |

## CI セーフティネット

`.github/workflows/release.yml` はタグプッシュ時に自動実行される:

1. `v*` タグのプッシュを検知
2. 同名の GitHub Release が既に存在するか確認
3. 存在しなければ CHANGELOG から自動生成（セーフティネット）
4. 存在すれば何もしない

手動で GitHub Release を作成してからプッシュするのが推奨フロー。
セーフティネットは「Release 作成忘れ」のみを救う。

## PM ハンドオフ

リリース後に PM への完了報告:

```markdown
## リリース完了報告

**バージョン**: v{{NEW_VERSION}}
**リリース日**: {{DATE}}

### 実施内容
{{CHANGELOG の内容}}

### GitHub Release
{{URL}}

### 次のアクション
- PM によるリリースノートの確認
- 本番環境へのデプロイ（該当する場合）
```

## 禁止事項

- タグの削除・巻き戻し（公開済みバージョンは不変）
- 同日に 2 回以上の minor バンプ
- patch レベルの変更での minor バンプ
- `--force` / `--force-with-lease` による force push
- リリースコミットに VERSION / plugin.json / CHANGELOG 以外の実装変更を混入

## 関連スキル

- `harness-review` -- リリース前にコードレビューを実施
- `harness-work` -- リリース後の次のタスクを実装
- `harness-plan` -- 次バージョンの計画を作成
- `x-announce` -- X (Twitter) へのリリース告知スレッド生成
- `harness-setup` -- mirror 同期やプラグイン設定のセットアップ

## 関連ルール

- `.claude/rules/versioning.md` -- SemVer 判定基準とバッチリリース推奨
- `.claude/rules/github-release.md` -- GitHub Release Notes フォーマット（英語）
- `.claude/rules/cc-update-policy.md` -- CC アプデ追従時の Feature Table 品質基準
