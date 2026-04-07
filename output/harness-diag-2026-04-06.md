# Harness Diagnosis: claude-code-harness

> 評価基準: [diagnosis-rubric](diagnosis-rubric.md)
> 診断日: 2026-04-06

## ハーネス構成サマリ

| 項目 | 現状 |
|------|------|
| CLAUDE.md | 246行 / ポインタ 7件（docs/, .claude/rules/, go/） / Feature Table 約130行がインライン |
| Permissions | allow 16件（プロジェクト） / deny 1件（`mcp__codex__*`のみ） / ユーザーレベル allow 4件 + deny なし |
| Hooks | **プラグイン**: 27イベント・計67フック（PreToolUse 4, PostToolUse 17, PermissionRequest 2, SessionStart 8, SubagentStart/Stop 各4, PreCompact 2, PostCompact 1, Elicitation/ElicitationResult 各1, UserPromptSubmit 5, Stop 4, 他多数）/ **ユーザーレベル**: PermissionRequest 1, PostToolUse 1, PostToolUseFailure 1, Stop 1, UserPromptSubmit 1, PreToolUse 9（SSH guard） |
| Skills | 計34件（ワークフロー: harness-work, harness-plan, harness-review, harness-release, breezing, harness-setup, harness-sync / セッション管理: session, session-init, session-memory, session-state, session-control / ユーティリティ: memory, ci, auth, crud, deploy, ui, agent-browser, allow1 / 情報系: vibecoder-guide, workflow-guide, principles / 生成: generate-slide, generate-video, notebookLM, x-announce, x-article / 統合: cc-update-review, cc-cursor-cc, claude-codex-upstream-update, gogcli-ops / テスト: zz-review-escape, zz-review-empty） |
| MCP | プロジェクトレベル: 未使用 / ユーザーレベル: 1接続（ccagi-tools） |
| Memory | SSOT: decisions.md（31 ADR）+ patterns.md（20パターン）/ プロジェクトMemory: 1エントリ / Agent Memory: 3エージェント分 |
| Agents | カスタム 3件（agents-v3/: worker, reviewer, scaffolder）+ team-composition.md |
| Plugins | 3件（claude-code-harness@marketplace — 自身, security-ops-check@directory, codex@openai-codex）。自身が67 Hook + 34スキル + 3エージェントを提供 |

## スコアサマリ

| カテゴリ | 指標 | スコア | 小計 |
|---------|------|--------|------|
| **A. 帯域効率** | A1 ⚠️ A2 — A3 ✅ A4 ⚠️ A5 ✅ | 5/8 | 63% |
| **B. 検証の堅牢性** | B1 ✅ B2 ✅ B3 ✅ B4 ✅ B5 ✅ | 10/10 | 100% |
| **C. 権限と信頼境界** | C1 ⚠️ C2 ⚠️ C3 ❌ C4 ⚠️ C5 ⚠️ | 4/10 | 40% |
| **D. 知識と記憶** | D1 ✅ D2 ⚠️ D3 ✅ D4 ✅ D5 ✅ | 9/10 | 90% |
| **E. 環境設計** | E1 ✅ E2 ✅ E3 ✅ E4 ✅ E5 ✅ | 10/10 | 100% |
| **総合** | | **38/48** | **79%** |

> 「—」（対象外）は分母から除外。A2はプロジェクトレベルでMCP未使用のため対象外。

### グレード

**A** (79%) — 基盤はしっかり。権限と信頼境界の強化でSランクに到達できる。

---

## 検出されたアンチパターン

### C3. エージェントが自分のルールを書き換えられないか ❌

**検出事実**: プロジェクトレベルの `.claude/settings.json` の `permissions.deny` には `mcp__codex__*` のみ。`Edit(.claude/settings*)` は含まれていない。ユーザーレベルの `~/.claude/settings.json` にもsettings.json の自己編集保護はない。エージェントは settings.json を自由に編集でき、deny ルールの追加・削除が可能な状態。

**影響**: permissions.deny に今後セキュリティルールを追加しても、エージェントがそのdenyを自ら解除できる。鍵のかかった金庫の鍵が机の上にある状態。特にこのプロジェクトは「ハーネスでハーネスを改善する」自己参照構造のため、エージェントが自分の評価基準を変更できる経路が開いていることの影響は通常以上に大きい。

**関連原則**: (->C-5) 報酬ハッキング, (->S-1) 評価と実装の権限分離

**改善案**:

```jsonc
// .claude/settings.json の permissions.deny に追加
{
  "permissions": {
    "deny": [
      "mcp__codex__*",
      "Edit(.claude/settings*)",
      "Write(.claude/settings*)"
    ]
  }
}
```

---

### C1. 品質のものさし自体が守られているか ⚠️

**検出事実**: PreToolUse Hook に agent タイプの品質ガード（Haiku による hardcoded secrets, TODO stubs, security vulnerabilities チェック）が組み込まれている。TypeScript ガードレールエンジン（`core/src/guardrails/tampering.ts`）が存在し、テスト改ざん防止の3層防御戦略が decisions.md (D9) に記録されている。ただし、`permissions.deny` にはテストファイル（`**/*.test.ts`）やlint設定（`.eslintrc*`, `biome.json`）、`tsconfig.json` の保護が含まれていない。

**影響**: ガードレールエンジンとHookの2層で保護されているが、settings.json レベルの宣言的deny（Hookより前段階でブロック）がない。プラグイン読み込み前や Hook 実行エラー時には保護が効かない隙間がある。

**関連原則**: (->C-5) 報酬ハッキング, (->S-1) 信頼境界を明示的に設計する

**改善案**:

既存のHook防御に加え、宣言的な第4層として:
```jsonc
// .claude/settings.json の permissions.deny に追加
"Edit(**/*.test.ts)",
"Edit(**/*.spec.ts)",
"Edit(tsconfig.json)",
"Edit(.eslintrc*)",
"Edit(biome.json)"
```

---

### C2. 取り返しのつかない操作が気軽にできない設計か ⚠️

**検出事実**: `permissions.deny` に不可逆操作（`git push --force`, `git reset --hard`, `rm -rf`）の制限がない。Worker エージェントは `isolation: worktree` で影響範囲が制限されており、パターン記録（P14）で「ホワイトリスト方式の rm バイパス」が定義済み。ただし Lead セッションでの不可逆操作ガードは存在しない。

**影響**: Worktree 外のメインコンテキスト（Lead セッション等）では不可逆操作に対するガードがない。善意の間違いで force push やハードリセットが実行される可能性がある。

**関連原則**: (->S-1.3) 最小権限で爆発半径を抑える, (->C-7) 校正盲

**改善案**:

```jsonc
// .claude/settings.json の permissions.deny に追加
"Bash(git push --force*)",
"Bash(git push -f*)",
"Bash(git push * --force*)",
"Bash(git reset --hard*)",
"Bash(rm -rf /*)",
"Bash(rm -rf ~*)"
```

---

### C4. 外の世界に影響する操作に歯止めがあるか ⚠️

**検出事実**: `permissions.allow` に `git push` は含まれていないため、デフォルトの Permission Mode で確認が求められる設計。ただし、ユーザーレベルの `skipDangerousModePermissionPrompt: true` により確認がスキップされる可能性がある。Worker エージェントは `permissionMode: bypassPermissions` が明示されている。ユーザーレベルの PreToolUse に SSH guard（9件）が設定されているのは良い設計。

**影響**: Permission Mode のデフォルト確認に依存しており、意図的な設計ではあるが、`skipDangerousModePermissionPrompt` との組み合わせで防護が弱まる。Worker（bypassPermissions）が誤って git push を実行した場合の歯止めがない。

**関連原則**: (->S-1.3) 最小権限, (->S-1.5) 安全側をデフォルトにする

**改善案**:

force push の deny（C2 で対応）に加え、通常の push は意図的な allowlist に含めずPermission Mode確認に委ねる現状の設計を維持。Worker エージェントの frontmatter に「外部影響操作は Lead が管理」の注記を追加することで設計意図を明文化。

---

### C5. 外から入ってくる情報を疑っているか ⚠️

**検出事実**: MCP 接続はユーザーレベルの ccagi-tools のみ（自作ツール）。Codex プラグイン経由の外部入力は companion スクリプトで構造化。Codex MCP は `deny: ["mcp__codex__*"]` でブロック済み。Elicitation ハンドラーが MCP 入力を処理。Memory Bridge が session-start/post-tool-use/stop で記憶を管理。

**影響**: 外部入力ソースは自作ツールに限定されており、不特定多数のソースからの注入リスクは低い。ただし Memory Bridge の記憶書き込み時に出所タグ付けが明示されていないため、将来の外部接続追加時に注意が必要。

**関連原則**: (->S-1.1) 記憶の出所と露出先を追跡, (->S-1.4) 防御は伝播停止の段階で評価する

**改善案**: 現時点では信頼ソース限定により実害は小さいため、優先度は低い。将来的に外部 MCP 接続を追加する際に Memory Bridge の出所タグ付け仕組みを検討。

---

### A1. CLAUDE.mdは地図になっているか ⚠️

**検出事実**: CLAUDE.md は246行。うち Feature Table が約130行（全体の53%）を占める。Feature Table は CC バージョンごとの対応状況を150件超記録しているが、毎セッションでこの全量が必要かは疑問。`docs/CLAUDE-feature-table.md`（100KB超）への詳細ポインタが151行目に存在。スタック宣言、開発フロー、テストコマンド等の毎セッション必要な情報はしっかり含まれている。

**影響**: 毎セッション130行の Feature Table がコンテキストに読み込まれるが、日常的な開発作業で参照される頻度は低い。エージェントの「机」の半分がカタログで埋まっている状態。

**関連原則**: (->C-1) コンテキスト帯域は有限でゼロサム, (->K-2.1) ポインタは百科事典より強い

**改善案**:

Feature Table をポインタ化:
```markdown
## Claude Code Feature Utilization

CC 2.1.92+ の機能活用状況は [docs/CLAUDE-feature-table.md](docs/CLAUDE-feature-table.md) を参照。
主要な活用: Agent Memory, Worktree isolation, PreCompact/PostCompact Hook, Agent Teams, Hooks conditional `if` field。
```

約100行削減により CLAUDE.md を約150行に圧縮。

---

### A4. エージェントの選択肢が不必要に広がっていないか ⚠️

**検出事実**: 34スキル中、`user-invocable: false` が14件、`disable-model-invocation: true` が2件（deploy, generate-video）。14件の非ユーザー呼出スキルは自動発動を適切に抑制。ただし残り20件が自動発動可能で、description 一覧がエージェントのメニューに展開される。`permissions.allow` の `Bash(cat *)`, `Bash(find *)`, `Bash(grep *)` は組み込みツール（Read, Glob, Grep）と機能重複。

**影響**: 自動発動抑制（14/34）と context:fork 分離（7スキル）は良い設計判断。ただし allow の重複コマンドがエージェントに不要な選択肢を提供している。

**関連原則**: (->V-1.3) 選択肢は推論前に絞る, (->C-1.1) コンテキストの能動的管理

**改善案**: 使用頻度の低い特定用途スキル（gogcli-ops, generate-video 等）に `disable-model-invocation: true` を追加。allow から Read/Glob/Grep で代替可能な `Bash(cat *)`, `Bash(find *)`, `Bash(grep *)` の除去を検討（スクリプト内利用がある場合は残す）。

---

### D2. 嘘を教えていないか ⚠️

**検出事実**: CLAUDE.md の167行目で `.claude/rules/changelog.md` を参照しているが、このファイルは存在しない（`ls` で確認済み）。`.claude/rules/` にはchangelogに関連するファイルが存在せず、最も近い内容は `github-release.md` の CHANGELOG セクション。それ以外の全参照パス（7件）は実在を確認。

**影響**: エージェントが CHANGELOG のフォーマットルールを参照しようとすると空振りする。致命的ではないが、CLAUDE.md の信頼性にギャップが生じる。

**関連原則**: (->K-2) 古い情報はノイズ, (->K-2.2) ドキュメントは構造化する

**改善案**:

CLAUDE.md 167行目を修正:
```markdown
// 現在
Details: [.claude/rules/changelog.md](.claude/rules/changelog.md)
// 修正案（2択）
// A. changelog.md を新規作成する
// B. 既存の github-release.md へ参照先を変更
Details: [.claude/rules/github-release.md](.claude/rules/github-release.md) (CHANGELOG セクション参照)
```

---

## 強み

### 1. 検証ループの完成度（B: 100%）

4層構造の検証設計が際立つ。(1) PreToolUse の TypeScript ガードレールエンジン（決定論的チェック）、(2) PreToolUse の Agent Hook（Haiku によるセキュリティ・品質チェック）、(3) PostToolUse の品質パック + 自動テスト実行（async）+ TDD順序検証、(4) PreCompact での WIP タスク警告。テスト改ざん防止は decisions.md D9 に「3層防御戦略」として根拠付きで設計されており、B5（同じ失敗の仕組み化）の好例。

### 2. 環境設計の一貫性（E: 100%）

Plan -> Work -> Review の3フェーズ分離、Worker/Reviewer のコンテキスト分離（Reviewer は `disallowedTools: [Write, Edit, Bash, Agent]` で Read-only を強制）、`context: fork` による7スキルの評価隔離。判断理由が CLAUDE.md とルールファイルの両方に記載されている。31件の ADR と20件のパターンが「なぜそうなっているか」を構造的に保持し、劣化対抗の基盤になっている。

### 3. Hook カバレッジの網羅性

27イベント・67フックは Claude Code の Hook API をほぼ全網羅。SessionStart の startup/resume 分岐、SubagentStart/Stop のエージェント別トラッキング、PreCompact/PostCompact のコンテキスト保護、Elicitation/ElicitationResult の MCP 対応、PermissionDenied の追跡、StopFailure のエラーキャプチャまで、セッションライフサイクルの全段階をカバーしている。

---

## Quick Wins — 今日できる改善

### 1. settings.json の自己保護（C3、5分）

最優先。他のdenyルール追加の前提条件。

```jsonc
// .claude/settings.json の permissions.deny を更新
{
  "permissions": {
    "deny": [
      "mcp__codex__*",
      "Edit(.claude/settings*)",
      "Write(.claude/settings*)"
    ]
  }
}
```

### 2. 不可逆操作の deny 追加（C2、5分）

```jsonc
// permissions.deny に追加
"Bash(git push --force*)",
"Bash(git push -f*)",
"Bash(git reset --hard*)"
```

### 3. テスト・lint設定の保護（C1、5分）

```jsonc
// permissions.deny に追加
"Edit(**/*.test.ts)",
"Edit(**/*.spec.ts)",
"Edit(tsconfig.json)",
"Edit(.eslintrc*)",
"Edit(biome.json)"
```

### 4. 存在しない changelog.md 参照の修正（D2、5分）

CLAUDE.md 167行目:
```markdown
Details: [.claude/rules/github-release.md](.claude/rules/github-release.md) (CHANGELOG セクション参照)
```

### 5. Feature Table のポインタ化（A1、15分）

CLAUDE.md の Feature Table（約130行）をポインタに圧縮:
```markdown
## Claude Code Feature Utilization

CC 2.1.92+ の機能活用状況は [docs/CLAUDE-feature-table.md](docs/CLAUDE-feature-table.md) を参照。
主要な活用: Agent Memory, Worktree isolation, PreCompact/PostCompact Hook, Agent Teams, Hooks conditional `if` field。
```

---

## 次のステップ — 中期的な改善

### 1. permissions.deny の体系的整備

Quick Wins の個別追加後、「設定ファイル保護」「品質ものさし保護」「不可逆操作保護」の3カテゴリで deny ルールを体系化。hooks.json、agents-v3/ のエージェント定義、core/ のガードレールエンジン自体も保護候補。

### 2. 低頻度スキルの自動発動抑制（A4）

34件のスキルを使用頻度で分類し、特定用途スキル（gogcli-ops, generate-video, notebookLM 等）に `disable-model-invocation: true` を追加。コアワークフローへの集中力を高める。

### 3. Memory Bridge の出所追跡（C5）

将来の外部 MCP 接続追加に備え、記憶書き込み時の出所タグ（source: mcp/codex/user/agent）付与の仕組みを設計。

---

## 総評

claude-code-harness は**検証の堅牢性（B: 100%）と環境設計（E: 100%）が満点**の成熟したハーネス設計。27イベント・67フックの Hook カバレッジ、TypeScript ガードレールエンジン、3エージェントの責務分離（Worker は実装・Reviewer は Read-only・Scaffolder は足場）、31件の ADR と20件のパターンによる判断の外部化 --- いずれもプロダクションレベルの設計蓄積を示している。

最大の改善領域は**権限と信頼境界（C: 40%）**。settings.json の自己保護（C3）が未設定であるため、他の deny ルールの信頼性基盤が不安定。特にこのプロジェクトは自己参照構造（ハーネスでハーネスを改善する）のため、エージェントが自分の評価基準を変更できる経路を閉じることが品質保証の土台となる。Quick Wins 1-3 を合わせて15分で C カテゴリを大幅に改善できる。

本診断はファイルベースの静的解析であり、Hook スクリプトの動作品質、エージェント間のハンドオフ実績、ガードレールエンジンのルール網羅性は検出範囲外である。
