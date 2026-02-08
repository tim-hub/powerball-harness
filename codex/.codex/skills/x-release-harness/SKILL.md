---
name: release-harness
description: "Harness リリース作業を自動化。CHANGELOG、バージョン、タグをポチッと一発。Use when user mentions harness release, harness version bump. Do NOT load for: general release discussions, other project releases."
description-en: "Automate Harness release. CHANGELOG, version, tag in one click. Use when user mentions harness release, harness version bump. Do NOT load for: general release discussions, other project releases."
description-ja: "Harness リリース作業を自動化。CHANGELOG、バージョン、タグをポチッと一発。Use when user mentions harness release, harness version bump. Do NOT load for: general release discussions, other project releases."
allowed-tools: ["Read", "Write", "Edit", "Bash"]
argument-hint: "[patch|minor|major]"
user-invocable: false
context: fork
---

# Release Harness Skill

Automates the claude-code-harness release process.

## Quick Reference

- "**Release a new harness version**" → `/release-harness`
- "**Bump patch version**" → `/release-harness patch`
- "**Create a minor release**" → `/release-harness minor`

---

## Execution Flow

### Step 1: Change Verification

Run in parallel:
1. `git status` - Check uncommitted changes
2. `git diff --stat` - List changed files
3. `git log --format="%h|%s|%an|%ad" --date=short -10` - Recent commit history (structured)

### Git log 拡張フラグの活用（CC 2.1.30+）

リリースノート生成時に構造化ログを活用します。

#### リリースノート用のコミット一覧

```bash
# 構造化フォーマットでコミット一覧取得
git log --format="%s" vPREV..HEAD

# マージコミットを除外（実質的な変更のみ）
git log --cherry-pick --no-merges --format="%s" vPREV..HEAD

# 詳細情報付き（作者・日付込み）
git log --format="%h|%s|%an|%ad" --date=short vPREV..HEAD
```

#### 主な活用場面

| 用途 | フラグ | 効果 |
|------|--------|------|
| **リリースノート生成** | `--format="%s"` | コミットメッセージのみ抽出 |
| **マージ除外** | `--cherry-pick --no-merges` | 実コミットのみでノート作成 |
| **詳細一覧** | `--format="%h\|%s\|%an\|%ad"` | 構造化された詳細情報 |
| **変更ファイル** | `--raw` | 影響範囲の把握 |

#### 出力例

```markdown
📝 リリース準備: v2.18.0

前回リリース（v2.17.10）からのコミット:

| Hash | Subject | Author | Date |
|------|---------|--------|------|
| a1b2c3d | feat: add git log flags | Alice | 2026-02-04 |
| e4f5g6h | docs: update CI docs | Bob | 2026-02-03 |
| i7j8k9l | fix: build script | Charlie | 2026-02-02 |

自動生成されたリリースノート（マージ除外）:
- feat: add git log flags
- docs: update CI docs
- fix: build script

→ CHANGELOG.md の下書きに使用
```

### Step 2: Version Determination

Check current version:
```bash
cat VERSION
```

Determine version based on changes ([Semantic Versioning](https://semver.org/)):
- **patch** (x.y.Z): Bug fixes, minor improvements
- **minor** (x.Y.0): New features (backward compatible)
- **major** (X.0.0): Breaking changes

Ask user: "What should the next version be? (e.g., 2.5.23)"

### Step 3: CHANGELOG Update (JP + EN)

> **⚠️ 重要**: ユーザー体験に影響する変更を中心に記載。内部修正は簡潔に。

Update both `CHANGELOG_ja.md` and `CHANGELOG.md`.

#### CHANGELOG 記載ルール

| 変更タイプ | 記載方法 |
|-----------|---------|
| **ユーザー体験に影響** | `🎯 What's Changed for You` + Before/After テーブル |
| **新機能追加** | `Added` セクションで簡潔に |
| **内部修正（CI/テスト/ドキュメント）** | `Internal` セクションで1行のみ |
| **バグ修正（ユーザー影響あり）** | `Fixed` セクション |
| **バグ修正（内部のみ）** | 省略または `Internal` に統合 |

#### テンプレート

```markdown
## [X.Y.Z] - YYYY-MM-DD

### 🎯 What's Changed for You

**ユーザー体験の変化を1行で説明**

| Before | After |
|--------|-------|
| 変更前の状態 | 変更後の状態 |

### Added

- 新機能の簡潔な説明

### Internal

- 内部修正の1行サマリー
```

#### Before/After テーブルのルール

- **体験が変わる変更のみ** Before/After を記載
- 内部修正（CI、テスト、リファクタリング）には不要
- 技術詳細ではなく **ユーザー視点の変化** を記載

#### 悪い例 vs 良い例

```markdown
❌ 悪い例（技術詳細すぎる）:
- **agents/*.md**: スキル参照を更新（`review` → `harness-review`）
- **CI: validate-plugin.sh** が Skills 移行後も正常動作するように修正

✅ 良い例（ユーザー視点）:
### Internal
- CI/テスト/ドキュメントを Skills 移行後の構造に更新
```

### Step 3.5: README Update Check

> Check if README needs update (JP/EN both)

### Step 4: Version File Update

```bash
echo "X.Y.Z" > VERSION
```

Also update `.claude-plugin/plugin.json`:
```json
"version": "X.Y.Z"
```

### Step 5: Commit and Tag

```bash
git add -A
git commit -m "chore: release vX.Y.Z"
git tag -a vX.Y.Z -m "Release vX.Y.Z"
```

### Step 6: Push

```bash
git push origin main
git push origin vX.Y.Z
```

### Step 7: GitHub Release (Optional)

```bash
gh release create vX.Y.Z \
  --title "vX.Y.Z - Title" \
  --notes "$(cat <<'EOF'
## 🎯 What's Changed for You

**ユーザー体験の変化を1行で**

| Before | After |
|--------|-------|
| 変更前 | 変更後 |

### Added / Changed / Fixed

- 簡潔な説明

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## GitHub Release Format

Follow `.claude/rules/github-release.md`:

```markdown
## 🎯 What's Changed for You

**One-line value description**

| Before | After |
|--------|-------|
| Previous state | New state |

### Added / Changed / Fixed

- Brief description

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

---

## Related Skills

- `verify` - Pre-release verification
