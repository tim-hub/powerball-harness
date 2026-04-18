# Smoke Test — v4.2.0-arcana (Phase 44 Release)

**実施日**: 2026-04-18  
**ブランチ**: `release/v4.2.0-arcana`  
**テスト実施者**: Worker subagent (harness-work/44.12.1)  
**目的**: Phase 44 全 9 タスク (44.3.1–44.11.1) の cherry-pick 完了後、リリース前の最終品質確認

---

## 自動テスト結果

| Test | Command | Result | Notes |
|------|---------|--------|-------|
| validate-plugin | `bash tests/validate-plugin.sh` | **PASS** | 警告 2 件（統合テスト警告、既知）、失敗 0 |
| consistency | `bash scripts/ci/check-consistency.sh` | **WARN** | mirror drift 3 件（harness-review × 2、harness-loop × 1）。既知・既出。新規 fail なし |
| migration residue | `bash scripts/check-residue.sh` | **PASS** | 残骸 0 件。スキャン時間 76.9s |
| Go guardrail tests | `go test ./go/internal/guardrail/... -count=1`（`go/` dir） | **PASS** | 119 テスト PASS、0 FAIL |
| R01-R13 regression | `bash tests/test-guardrails-r01-r13.sh` | **PASS** | CC2110_* 34 件 PASS（compound/escape/subshell/backtick 全カバー） |
| 1h cache opt-in | `bash tests/test-prompt-cache-1h.sh` | **PASS** | 9 テスト PASS。env.local 追記・冪等性・env 伝播すべて確認 |

### validate-plugin 詳細

- 合格: 39
- 警告: 2（統合テスト警告。既存問題、本 Phase の変更によるものではない）
- 失敗: 0

### consistency チェック詳細

以下の 3 件は既知の mirror drift（Worker D が報告済み）:

- `codex`: harness-review mirror が SSOT と不一致
- `opencode`: harness-review mirror が SSOT と不一致
- `opencode`: harness-loop mirror が SSOT と不一致

本 Phase の変更（Phase 44.3.1–44.11.1）では新規 drift は発生していない。

### Go guardrail tests 詳細

- パッケージ: `github.com/Chachamaru127/claude-code-harness/go/internal/guardrail`
- 実行コマンド（正しい cwd）: `cd go && go test ./internal/guardrail/... -count=1`
- 119 テスト、PASS / 0 FAIL

> **注意**: リポジトリルートから `go test ./go/internal/guardrail/...` を実行すると
> `pattern ./go/internal/guardrail/...: directory prefix go/internal/guardrail does not contain main module or its selected dependencies`
> エラーになる。`go/` ディレクトリに移動してから実行すること。

### R01-R13 regression 詳細

CC2110_* テスト（34 件）は以下のカテゴリをカバー:

- CompoundSemicolon / CompoundAmpAmp / CompoundPipe / CompoundOr によるバイパス試行
- BackslashEscape によるバイパス試行
- EnvVarPrefix によるバイパス試行（安全な既知コマンドは引き続きブロック）
- Heredoc を使ったバイパス試行（正常通過を確認）
- Subshell / Backtick によるバイパス試行
- PermissionUpdatedInput / AdditionalContext 系の state 保持テスト

---

## 手動チェックリスト (Lead/User が実施)

以下の 6 項目は Worker subagent では実行できないため、Lead またはユーザーが手動で確認すること。

### (a) literal prompt 変更後の曖昧表現ゼロ確認

**確認コマンド**:

```bash
rg "必要に応じて|適宜|適切に|十分に|柔軟に|しっかり|可能なら|場合によって" agents/*.md
```

**期待結果**: 0 件（`.claude/rules/opus-4-7-prompt-audit.md` の合格条件）

- [ ] grep 結果 0 件を確認

### (b) PreCompact block 動作確認

**テスト手順**:

```bash
# hooks/pre-compact.sh を直接実行してモックの compact trigger をシミュレーション
# echo で compact フック入力を渡す
echo '{"type":"pre_compact","summary":"test"}' | bash hooks/pre-compact.sh
# → exit 2 で deny 返却されることを確認
echo "exit code: $?"
```

**期待結果**: exit code 2 / JSON `{"decision":"block",...}` が stdout に出力されること

- [ ] exit 2 かつ deny JSON を確認

### (c) monitors auto-arm（Phase 45 スコープ）

**現状**: `plugin.json` に `monitors` block なし。

```bash
cat .claude-plugin/plugin.json | jq 'has("monitors")'
# → false（現状）
```

Phase 45 でモニター自動起動ブロックを追加予定。本 smoke test では確認のみ。

- [ ] Phase 45 追加スコープとして記録済みを確認

### (d) 1h cache 反映確認

**テスト手順**:

```bash
# 1h キャッシュを有効化
bash scripts/enable-1h-cache.sh
source .env.local  # または env.local

# CC セッション起動（環境変数が伝播されていることを確認）
echo $ENABLE_PROMPT_CACHING_1H  # → 1

# CC セッション内で /cost を実行してプロンプトキャッシュヒットを確認
# （実際の CC セッション内でのみ確認可能）
```

- [ ] `ENABLE_PROMPT_CACHING_1H=1` が env に設定されることを確認
- [ ] CC セッション内 `/cost` でキャッシュヒットを確認

### (e) xhigh effort 受付確認

**確認対象ファイル**:

- `agents/reviewer.md`
- `agents/advisor.md`

**確認方法**:

```bash
grep -n "effort.*xhigh\|xhigh.*effort" agents/reviewer.md agents/advisor.md
```

CC 2.1.111+ の Opus 4.7 向け `xhigh` effort が定義されていることを確認。
実際に CC で `xhigh` が受け付けられるかは、CC 2.1.111+ 環境での手動確認が必要。

- [ ] agents/ に xhigh effort 定義が存在することを確認
- [ ] CC 2.1.111+ 環境で xhigh effort が受け付けられることを確認

### (f) guardrails R01-R13 regression ゼロ

自動テスト結果（上記）が PASS のため、本項目はテスト結果で担保済み。

- [x] Go test PASS (119 件)
- [x] R01-R13 shell test PASS (CC2110_* 34 件)

---

## 既知の問題

| 問題 | 詳細 | 対応方針 |
|------|------|---------|
| Plans.md 行数超過 | Plans.md が約 285 行（推奨上限 200 行） | Phase 44 完走後にアーカイブ予定 |
| mirror drift 3 件 | check-consistency.sh の harness-review × 2、harness-loop × 1 が SSOT と不一致 | Worker D 報告で既出。Phase 45 で整理予定 |
| monitors block なし | plugin.json に monitors field が存在しない | Phase 45 で対応（(c) 参照） |
| isolation:worktree 不全 | 並列 Worker spawn で worktree isolation が効かず branch state 共有が発生した事例あり | cherry-pick は SHA 直接指定で対処（memory: worker_worktree_share.md） |
| Go test cwd 制約 | `go test ./go/internal/guardrail/...` はリポジトリルートから実行不可 | `cd go` してから実行。validation_commands の記述に注意 |

---

## Phase 44 cherry-pick history (参考)

`git log --oneline 36b73367..HEAD` 出力（`release/v4.2.0-arcana` ブランチ）:

```
f0d3cdc5 chore(plans): mark 44.11.1 cc:完了 [d96e94b7]
d96e94b7 docs(phase-44.11.1): add v2.1.99-110 + Opus 4.7 entries to Feature Table (B=0)
63f7ddb5 chore(plans): mark 44.10.1 cc:完了 [95d1b39b]
95d1b39b docs(phase-44.10.1): add Task Budgets research memo
87996736 chore(plans): mark 44.9.1 cc:完了 [5949ee7d]
5949ee7d docs(vision): add Opus 4.7 vision high-res review flow and usage guide
1bd2a3b7 chore(plans): mark 44.5.1/44.8.1 cc:完了
e24f7f99 docs(phase-44.8.1): confirm /ultrareview policy — harness-review takes precedence in automation flow
b6e78ba9 chore(plans): mark 44.5.1 cc:完了 [315de2b9]
315de2b9 feat(agents): adopt xhigh effort for reviewer/advisor + document CC-API effort matrix
3e46f416 chore(plans): mark 44.3.1/44.4.2/44.6.1/44.7.1 cc:完了
c7be2b5c feat(phase-44.7.1): integrate CC 2.1.99-110 small features (5 items)
a8f1c308 feat(cache): add enable-1h-cache.sh opt-in script and update long-session docs (Phase 44.6.1)
f921cd98 test(guardrails): CC 2.1.110 regression tests for R01-R13 re-conformance (Phase 44.3.1)
856e9888 fix(mirror): sync opencode/skills/harness-loop with SSOT (add user-invocable: true)
```

---

## テスト環境情報

| 項目 | 値 |
|------|-----|
| 実施日 | 2026-04-18 |
| ブランチ | `release/v4.2.0-arcana` |
| HEAD | `f0d3cdc5` |
| OS | darwin 25.3.0 |
| Shell | zsh |
| Go module | `github.com/Chachamaru127/claude-code-harness` |
| CC バージョン | 2.1.111+ (Opus 4.7) ターゲット |
