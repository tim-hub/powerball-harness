# Harness Diagnosis: claude-code-harness

> 評価基準: [diagnosis-rubric](diagnosis-rubric.md)
> 診断日: 2026-04-07

## ハーネス構成サマリ

| 項目 | 現状 |
|------|------|
| CLAUDE.md | 135行 / ポインタ 6件（feature-table, changelog, commands, skill-catalog, SPEC.md, DESIGN.md） / インライン手順 3件（commit style, version mgmt, dev flow） |
| Permissions | allow 16件（`.claude/settings.json`） / deny 1件（project `.claude/settings.json`: `mcp__codex__*`のみ） / deny 16件（plugin `.claude-plugin/settings.json`: sudo, rm -rf, DB, secrets等） / ask 15件（plugin: git force, reset, rebase, npm install等） |
| Hooks | PreToolUse 3グループ / PostToolUse 5グループ / SessionStart 2 / Stop 1 / PreCompact 1 / PostCompact 1 / Elicitation 2 / SubagentStart/Stop 各4 / UserPromptSubmit 1 / PermissionRequest 2 / Notification 1 / ConfigChange 1 / CwdChanged 1 / FileChanged 1 / WorktreeCreate/Remove 各1 / TaskCompleted 1 / TaskCreated 1 / TeammateIdle 1 / InstructionsLoaded 1 / SessionEnd 1 / PermissionDenied 1 / StopFailure 1 / PostToolUseFailure 1 — **計27イベント種別カバー** |
| Skills | 計33件（ワークフロー型: harness-work, harness-plan, harness-review, harness-release, harness-setup, breezing / 辞書型: principles, vibecoder-guide, workflow-guide / ユーティリティ: memory, session系5件, ci, auth, crud, deploy, ui, x-announce, x-article 等） |
| MCP | ユーザーレベルで harness MCP（harness_mem_*, harness_session_*, harness_workflow_*, harness_ast_*, harness_lsp_*）, ccagi-tools, datadog, pencil, stitch, Notion, Slack, figma |
| Memory | decisions.md（31エントリ D1-D31） / patterns.md（20エントリ P1-P20） / session-log.md / codex-learnings.md / archive/ 多数 |
| Agents | カスタム 4件（worker, reviewer, scaffolder, team-composition） |
| Plugins | 2件（security-ops-check@security-ops-check-marketplace, codex@openai-codex）。security-ops-check はセキュリティスキル提供、codex は Codex CLI 統合 |

## スコアサマリ

| カテゴリ | 指標 | スコア | 小計 |
|---------|------|--------|------|
| **A. 帯域効率** | A1 ✅ A2 ⚠️ A3 ✅ A4 ⚠️ A5 ✅ | 8/10 | 80% |
| **B. 検証の堅牢性** | B1 ✅ B2 ✅ B3 ✅ B4 ✅ B5 ✅ | 10/10 | 100% |
| **C. 権限と信頼境界** | C1 ⚠️ C2 ✅ C3 ❌ C4 ✅ C5 ✅ | 7/10 | 70% |
| **D. 知識と記憶** | D1 ✅ D2 ⚠️ D3 ✅ D4 ✅ D5 ✅ | 9/10 | 90% |
| **E. 環境設計** | E1 ✅ E2 ✅ E3 ✅ E4 ✅ E5 ✅ | 10/10 | 100% |
| **総合** | | **44/50** | **88%** |

> 「—」（対象外）は分母から除外。スコアはプロジェクトの規模・性質に応じた適用可能指標のみで算出。

### グレード

| グレード | 範囲 | 意味 |
|---------|------|------|
| **S** | 90%+ | ハーネス設計が成熟。微調整のフェーズ |
| **A** | 75-89% | 基盤はしっかり。いくつかの強化ポイントがある |
| **B** | 60-74% | 基本構造はあるが、構造的な穴がある |
| **C** | 40-59% | 重要な設計パターンが未導入。改善効果が大きい |
| **D** | <40% | ハーネス設計の初期段階。Quick Winsから始めよう |

**このプロジェクトのグレード: A（88%）**

## 検出されたアンチパターン

### C3. エージェントが自分のルールを書き換えられないか ❌

**検出事実**: CLAUDE.md の Permission Boundaries セクション（L98-109）には「`.claude-plugin/settings*`, `.claude/settings*` が deny によりハードブロック」と記述されている。しかし実際の設定ファイルを確認すると:

- `.claude/settings.json`（プロジェクトレベル）: deny は `mcp__codex__*` の1件のみ。settings.json 自体の保護は**存在しない**
- `.claude-plugin/settings.json`（プラグインレベル）: deny に `sudo`, `rm -rf`, DB接続, 秘密ファイル等はあるが、**settings.json の自己保護は含まれていない**
- ユーザーレベル `~/.claude/settings.json`: deny 記載なし（allow のみ）

つまり、CLAUDE.md が「ハードブロック」と宣言している settings.json の自己書き換え防止が、**どのレベルの settings.json にも実際には設定されていない**。エージェントは permissions.deny を含む settings.json を自由に編集でき、自分の制約を解除できる状態にある。

**影響**: deny ルールの信頼性が根本から崩壊する。エージェントが善意の判断で deny を緩和する（「この操作に必要なので deny を一時的に外します」）リスクがある。他の全ての deny ルールの実効性が、この1点の欠落で担保されない。

**関連原則**: (→C-5) 報酬ハッキング, (→S-1) 評価と実装の権限分離

**改善案**:

`.claude/settings.json` の `permissions.deny` に以下を追加:

```jsonc
{
  "permissions": {
    "deny": [
      "mcp__codex__*",
      "Edit(.claude/settings*)",
      "Write(.claude/settings*)",
      "Edit(.claude-plugin/settings*)",
      "Write(.claude-plugin/settings*)"
    ]
  }
}
```

---

### C1. 品質のものさし自体が守られているか ⚠️

**検出事実**: CLAUDE.md（L104）には「`.eslintrc*`, `eslint.config.*`, `biome.json`, `tsconfig*.json` が deny によりハードブロック」と明記されている。しかし:

- `.claude/settings.json`: これらの保護は deny に含まれていない
- `.claude-plugin/settings.json`: 同様に含まれていない

`core/tsconfig.json` や lint 設定はプロジェクトに存在し、テスト品質の基盤となっているが、エージェントが編集可能な状態。ただし `.claude/settings.json` の allow リストが `Edit(*)` で明示的に全ファイル編集を許可しているため、deny がなければ無制限に編集できる。

同様に `.github/workflows/*` の保護も CLAUDE.md に記載されているが settings.json に反映されていない。

**影響**: テスト設定（tsconfig の strict mode 等）をエージェントが緩和することで、テストが通りやすくなるが品質が下がる「ものさしの改ざん」リスクがある。ただし、agent hook による PreToolUse の品質チェック（hooks.json L25-28）と PostToolUse のコードレビュー（hooks.json L430-438）が補完的に機能しているため、完全に無防備ではない。

**関連原則**: (→C-5) 報酬ハッキング, (→S-1) 信頼境界を明示的に設計する

**改善案**:

`.claude/settings.json` の `permissions.deny` に追加:

```jsonc
"Edit(.eslintrc*)",
"Edit(eslint.config.*)",
"Edit(biome.json)",
"Edit(tsconfig*.json)",
"Write(.eslintrc*)",
"Write(eslint.config.*)",
"Write(biome.json)",
"Write(tsconfig*.json)",
"Edit(.github/workflows/*)",
"Write(.github/workflows/*)"
```

---

### D2. 嘘を教えていないか ⚠️

**検出事実**:

1. **VERSION / plugin.json 不一致**: `VERSION` = `3.17.1` / `.claude-plugin/plugin.json` の version = `3.17.0`。バージョンドリフトが発生している

2. **CLAUDE.md の changelog.md 参照が壊れている**: L32 に `[.claude/rules/changelog.md](.claude/rules/changelog.md)` へのリンクがあるが、`.claude/rules/changelog.md` は存在しない。14個の rules ファイルのうち changelog.md だけが欠落

3. **CLAUDE.md の Permission Boundaries が実態と乖離**: 上記 C3/C1 で指摘の通り、CLAUDE.md が「deny でハードブロック」と宣言している項目が実際の settings.json に反映されていない。エージェントがこの記述を信じて「settings.json が守ってくれている」と判断すると、実際には保護されていない操作を安全だと誤認する

**影響**: バージョン不一致は軽微だが、CLAUDE.md のリンク切れと Permission Boundaries の虚偽記載は、エージェントの判断を誤らせる。特に Permission Boundaries の乖離は C3 の問題と複合的に作用する

**関連原則**: (→K-2) 古い情報はノイズ, (→K-2.2) ドキュメントは構造化する

**改善案**:

1. VERSION 同期: `./scripts/sync-version.sh` を実行
2. changelog.md のリンクを修正（ファイルを作成するか、github-release.md 等の代替先にリンクを更新）
3. Permission Boundaries セクションは C3/C1 の改善後に実態と一致させる

---

### A2. ツール接続は帯域コストに見合っているか ⚠️

**検出事実**: ユーザーレベルで接続されている MCP サーバーが多数ある（harness, ccagi-tools, datadog, pencil, stitch, Notion, Slack, figma）。これはユーザーレベルの設定であり、全プロジェクトで常にツール定義がコンテキストに常駐する。

このプロジェクト（claude-code-harness 自身の開発）で必要なのは主に harness MCP と ccagi-tools。datadog, pencil, stitch, Notion, Slack, figma はこのプロジェクトの開発作業では使用頻度が低い可能性がある。

ただし、ユーザーが複数プロジェクトを横断して作業する場合、ユーザーレベルの MCP 設定は利便性のためのトレードオフとして正当な判断。

**影響**: 各 MCP サーバーのツール定義がコンテキストを消費し、エージェントの選択肢メニューを膨張させている。特に pencil (7ツール), stitch (12ツール), datadog (18ツール) は合計で相当のトークン数を占有

**関連原則**: (→C-1) CLIは帯域ゼロで実行できる

**改善案**: プロジェクトレベルの `.mcp.json` でこのプロジェクト固有に必要な MCP のみを有効化し、ユーザーレベルの不要な MCP をプロジェクト単位で無効化する設計も検討可能。ただし現状のユーザーレベル設定はマルチプロジェクト運用の利便性として理解できるため、優先度は低い。

---

### A4. エージェントの選択肢が不必要に広がっていないか ⚠️

**検出事実**: スキル33件が全て `user-invocable: true`（明示的に false にしているものなし）で、自動発動の対象。description にトリガーフレーズと除外条件が丁寧に記載されている点は良いが、33件のスキルメニューはエージェントの判断負荷としてはかなり大きい。

特に以下は自動発動を抑制しても良い候補:
- `zz-review-escape` / `zz-review-empty`: テンポラリスキル
- `cc-cursor-cc`: Cursor 連携固有
- `generate-video`, `generate-slide`, `notebookLM`: 特定用途
- `gogcli-ops`: Google Workspace 操作

**影響**: エージェントが毎回のリクエストで33件のスキルを走査し、最適なスキルを選択する判断コストが発生。大半のセッションで使用されないスキルが常時メニューに並んでいる

**関連原則**: (→V-1.3) 選択肢は推論前に絞る, (→C-1.1) コンテキストの能動的管理

**改善案**: 低頻度スキルに `disable-model-invocation: true` を設定し、明示的な `/skill-name` でのみ起動可能にする。ワークフローの核（harness-work, harness-review, harness-plan, harness-release, harness-setup, breezing, memory）以外は検討対象。

---

## 強み

1. **Hook 設計の網羅性と多層性**: 27イベント種別をカバーし、PreToolUse に agent hook（haiku による品質チェック）、PostToolUse に command + agent の二層検証、PreCompact/PostCompact でコンテキスト保全、Stop フックで WIP タスクゲートを実装。CC の Hook API を最大限に活用した設計は模範的

2. **生成と評価の分離**: `harness-review` スキルに `context: fork` が設定され、レビューが生成と物理的に別コンテキストで実行される。reviewer エージェントは `disallowedTools: [Write, Edit, Bash, Agent]` で Read-only に制限。迎合性を構造的に排除している

3. **SSOT としての Memory 設計**: decisions.md（31件）と patterns.md（20件）がインデックス付きで構造化され、session-log.md との3層分離が機能している。過去の意思決定が検索可能で、セッション間の引き継ぎ品質が高い

## Quick Wins — 今日できる改善

### 1. settings.json の自己保護を追加（C3対応、5分）

最も爆発半径が大きい欠落。以下を `.claude/settings.json` の `permissions.deny` に追加:

```jsonc
// .claude/settings.json
{
  "permissions": {
    "deny": [
      "mcp__codex__*",
      "Edit(.claude/settings*)",
      "Write(.claude/settings*)",
      "Edit(.claude-plugin/settings*)",
      "Write(.claude-plugin/settings*)"
    ]
  }
}
```

### 2. 品質基準ファイルの保護を追加（C1対応、5分）

CLAUDE.md で宣言済みの保護を実際に反映する。上記の deny に追加:

```jsonc
"Edit(.eslintrc*)",
"Edit(eslint.config.*)",
"Edit(biome.json)",
"Edit(tsconfig*.json)",
"Edit(.github/workflows/*)",
"Write(.eslintrc*)",
"Write(eslint.config.*)",
"Write(biome.json)",
"Write(tsconfig*.json)",
"Write(.github/workflows/*)"
```

### 3. VERSION / plugin.json の同期（D2対応、1分）

```bash
./scripts/sync-version.sh
```

### 4. changelog.md リンクの修正（D2対応、2分）

CLAUDE.md L32 の `[.claude/rules/changelog.md]` を、存在するファイル（例: `.claude/rules/github-release.md`）に更新するか、changelog.md ファイルを作成する。

### 5. テンポラリスキルの自動発動抑制（A4対応、5分）

`skills/zz-review-escape/SKILL.md` と `skills/zz-review-empty/SKILL.md` の frontmatter に追加:

```yaml
disable-model-invocation: true
```

## 次のステップ — 中期的な改善

### 1. CLAUDE.md Permission Boundaries の実態同期

C3/C1 の改善完了後、CLAUDE.md の Permission Boundaries テーブルを実際の deny/ask 設定と完全に一致させる。現在「deny でハードブロック」と書かれている `git push --force` と `git reset --hard` は、実際には `.claude-plugin/settings.json` の `ask`（確認付き許可）に設定されている。表記を実態に合わせるか、deny に昇格させるかを判断する。

### 2. プロジェクトレベル MCP 設定の最適化

`.mcp.json` を作成し、このプロジェクト固有に必要な MCP（harness, ccagi-tools）のみを明示的に有効化する構成を検討する。ユーザーレベルの汎用 MCP はそのまま維持しつつ、プロジェクト固有の帯域最適化を図る。

### 3. スキルの自動発動スコープ整理

33件のスキルを「常時自動発動」「明示起動のみ」に分類し、低頻度スキルに `disable-model-invocation: true` を適用する。目安として、週1回未満の使用頻度のスキルは明示起動のみに移行。

## 総評

claude-code-harness は Hook 設計、評価分離、Memory 構造の3点で卓越した成熟度を示している。特に27イベント種別の Hook カバレッジと agent hook による二層検証は、CC のプラグインエコシステムでもトップクラスの設計。グレード A（88%）はその実力を正確に反映している。

最も効果的な改善領域は **C（権限と信頼境界）** であり、具体的には settings.json の自己保護（C3）と品質基準ファイルの保護（C1）の2点。CLAUDE.md には既にこれらの保護方針が明記されているため、設定ファイルへの反映のみで解決する。意図は正しく、実装の反映が追いついていない状態。

本診断はファイルベースの静的解析であり、実際のワークフロー品質（Hook の実行時動作、agent hook の判定精度、Breezing のチーム協調品質等）は検出範囲外である点に留意されたい。
