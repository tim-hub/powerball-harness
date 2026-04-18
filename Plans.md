# Claude Code Harness — Plans.md

最終アーカイブ: 2026-04-17（Phase 37 + 41 + 42 + 43 → `.claude/memory/archive/Plans-2026-04-17-phase37-41-42-43.md`）
前回アーカイブ: 2026-04-15（Phase 25-36 + 38 → `.claude/memory/archive/Plans-2026-04-15-phase25-36-38.md`）

---

## 📦 アーカイブ

完了済み Phase は以下のファイルへ切り出し済み（git history にも残存）:

- [Phase 37 + 41 + 42 + 43](.claude/memory/archive/Plans-2026-04-17-phase37-41-42-43.md) — Hokage 完全体 / Long-Running Harness / Go hot-path migration / Advisor Strategy
- [Phase 39 + 40 + 41.0](.claude/memory/archive/Plans-2026-04-15-phase39-40-41.0.md) — レビュー体験改善 / Migration Residue Scanner / Long-Running Harness Spike

---

## Phase 44: Opus 4.7 / Claude Code 2.1.99-2.1.110 追従 — "Arcana" リリース

作成日: 2026-04-17
目的: Anthropic Claude Opus 4.7 (`claude-opus-4-7`) のリリースと Claude Code 2.1.99〜2.1.110 の機能群に Harness を完璧に追従させる。特に Opus 4.7 の "literal instruction following" と新 `xhigh` effort level、CC の `PreCompact` hook / `monitors` manifest / 1-hour prompt caching を Harness の Plan-Work-Review 全体に組み込み、v4.2.0 としてリリースする。
前提: v4.1.1 "post-Hokage" の Go native 実装が稼働、Phase 43 (Advisor Strategy) 完了済み

### 背景 (Why this phase exists)

Anthropic は 2026-04 に Claude Opus 4.7 をリリースした。主要な挙動変化:

1. **Instruction following が literal 化**: 「必要に応じて」「いい感じに」など曖昧な指示で Opus 4.6 時代は暗黙に補完されていた挙動が、4.7 では書かれたとおりにしか動かない。Anthropic 自身が "users need to re-tune prompts and harnesses" と明言している
2. **新 effort level `xhigh`**: `high` と `max` の間の強度。既定が `xhigh` になったプランもある
3. **Task Budgets (public beta)**: developer が token spend を宣言できる
4. **Tokenizer 変更 (1.0-1.35×)**: 同入力でも消費量が変動
5. **Vision 2576px / 3.75MP**: 3 倍以上の高解像度画像対応
6. **File-system memory の強化**: long multi-session での memory 利用が上達

Claude Code 側も v2.1.98 (Monitor ツール) 以降、v2.1.101 の `/team-onboarding` / settings resilience、v2.1.105 の **`PreCompact` hook** / **`monitors` manifest**、v2.1.108 の `ENABLE_PROMPT_CACHING_1H` / `/recap`、v2.1.110 の `PermissionRequest updatedInput` deny 再チェックなど、Harness の長時間実行・ガードレール・コスト最適化に直結する改善が連続で入った。

現行 Harness は依然として Opus 4.6 / CC 2.1.98 の世界観で設計されており、Feature Table は v2.1.98 で止まっている。放置すると (a) 既存プロンプトの誤作動、(b) 長時間 Worker の compaction 中断、(c) guardrails 迂回の発生、(d) トークンコスト増、が発生する。

### 設計方針

- **`.claude/rules/cc-update-policy.md` の A/B/C 分類を厳守**: Feature Table 追加は「実装あり (A)」または「CC 自動継承 (C)」でのみ許可。「書いただけ (B)」は禁止
- **Opus 4.7 の literal 化対応を最優先**: prompt re-tune を 44.1 直後に着手し、既存機能の回帰リスクを先に潰す
- **PreCompact / monitors / PermissionRequest fix は早期実装**: 長時間ワーカーの安定性とセキュリティに直結
- **`xhigh` effort は CC skill frontmatter の実効値を確認してから採用**: API の effort と CC の effort の対応関係が自明でないため、段階投入
- **Task Budgets は public beta のため調査のみとし、本採用は次サイクル**: Harness 側の budget 管理は既存 `max_consults` / `/cost` に任せる
- **既存の `bypassPermissions` は Auto Mode に切り替えない**: Auto Mode は Opus 4.7 Max user 向けで挙動差が読めないため、opt-in 扱いのまま v4.2 では維持

### 優先度マトリクス

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 44.1 | Impact assessment + A/B/C 分類表生成 | 1 | なし |
| **Required** | 44.2 | CC 2.1.105 `PreCompact` hook + `monitors` manifest 統合 | 2 | 44.1 |
| **Required** | 44.3 | Guardrails R01-R13 再検証 (`PermissionRequest updatedInput` deny, `additionalContext` persist) | 1 | 44.1 |
| **Required** | 44.4 | Opus 4.7 literal prompt re-tune (Worker / Reviewer / Advisor / Scaffolder + 5動詞スキル冒頭) | 2 | 44.1 |
| **Required** | 44.5 | `xhigh` effort 採用判定 + skill/agent `effort` frontmatter 更新 | 1 | 44.4 |
| **Recommended** | 44.6 | `ENABLE_PROMPT_CACHING_1H` opt-in + Breezing/harness-loop でのコスト削減 | 1 | 44.2 |
| **Recommended** | 44.7 | `EnterWorktree path` / `/recap` / `/undo` / skill mid-message fix (v2.1.105, v2.1.108, v2.1.110) の Harness 導入 | 1 | 44.2 |
| **Recommended** | 44.8 | `/ultrareview` と `/harness-review` の連携方針確定 (並立 or 委譲) + 実装 | 1 | 44.4 |
| **Optional** | 44.9 | Opus 4.7 Vision 2576px 対応 (harness-review の高解像度画像フロー) | 1 | 44.4 |
| **Optional** | 44.10 | Task Budgets (public beta) 調査メモ + 将来採用の判断材料整理 | 1 | 44.5 |
| **Required** | 44.11 | Feature Table v2.1.99-2.1.110 + Opus 4.7 セクション追加 (A/B/C 分類明記) | 1 | 44.2, 44.3, 44.4, 44.5 |
| **Required** | 44.12 | Smoke test + CHANGELOG v4.2.0 + リリース | 2 | 44.2-44.11 |

合計: **15 タスク**（Required 9 / Recommended 3 / Optional 2 / Release 1）

### 完成基準 (Definition of Done — Phase 44 全体)

| # | 基準 | 検証方法 | 必須/推奨 |
|---|------|---------|----------|
| 1 | `docs/CLAUDE-feature-table.md` に v2.1.99-v2.1.110 の全主要エントリが追加され、各行に A/B/C 分類が明記される | `tests/validate-plugin.sh` + grep で "A:" / "C:" がエントリ数分以上あること | 必須 |
| 2 | Opus 4.7 対応として agents/worker.md, reviewer.md, advisor.md, scaffolder.md, skills/harness-{work,review,plan,release,setup}/SKILL.md の曖昧表現が literal に書き直される | `.claude/rules/opus-4-7-prompt-audit.md` (新設) の checklist 全項目 PASS | 必須 |
| 3 | `.claude-plugin/hooks.json` に `PreCompact` エントリが追加され、長時間 Worker の意図せぬ compaction を block する仕組みが動く | `tests/test-pre-compact-hook.sh` (新設) PASS | 必須 |
| 4 | `.claude-plugin/plugin.json` に `monitors` manifest 項目が追加され、session 起動 / skill invoke で background monitor が auto-arm される | plugin validate PASS + `go test ./...` で monitor registry テスト PASS | 必須 |
| 5 | Guardrails (R01-R13) が `PermissionRequest updatedInput` と `PreToolUse additionalContext` の 2.1.110 仕様に再適合し、既存 deny chain を回帰なく維持する | `go/internal/hookhandler/*_test.go` の拡張 PASS + `tests/test-guardrails-r01-r13.sh` PASS | 必須 |
| 6 | `harness-work`, `harness-review`, `harness-plan` の skill frontmatter に `xhigh` effort を採用した場合の挙動差が smoke test で検証される (採用しない場合は rationale を docs に明記) | `docs/effort-level-policy.md` (新設 or 更新) + smoke test 記録 | 必須 |
| 7 | `ENABLE_PROMPT_CACHING_1H` を Breezing / harness-loop 起動時に opt-in できる仕組みが動き、`/cost` のキャッシュヒット率改善が確認できる | `scripts/enable-1h-cache.sh` (新設) + `tests/test-prompt-cache-1h.sh` PASS + 手動 `/cost` ログ | 推奨 |
| 8 | `/ultrareview` との関係が文書化され、`harness-review` が並立 or 委譲のどちらかを明示的に採用する | `docs/ultrareview-policy.md` (新設) + `skills/harness-review/SKILL.md` 更新 | 推奨 |
| 9 | `Plugin skills via "skills": ["./"]` + `disable-model-invocation: true` の mid-message 発火バグ修正 (v2.1.110) に対応した skill 実装が動く | 既存 skill 全てで `/` 呼び出しテスト PASS | 必須 |
| 10 | CHANGELOG v4.2.0 に GitHub Release rules (`.claude/rules/github-release.md`) 準拠の「今まで / 今後」エントリが記述され、CC 統合部分は「CC のアプデ → Harness での活用」形式で書かれる | CHANGELOG lint + マニュアルレビュー | 必須 |
| 11 | `VERSION` / `.claude-plugin/plugin.json` / `harness.toml` の 3 点が `4.2.0` に同期する | `scripts/sync-version.sh --check` PASS | 必須 |
| 12 | 既存の validate-plugin.sh / check-consistency.sh / Migration residue (check-residue.sh) が全て PASS | CI gate 全 PASS | 必須 |

---

### Phase 44.1: Impact assessment + A/B/C 分類表 [P0]

Purpose: Opus 4.7 + CC 2.1.99-110 の全変更点を Harness への影響度で分類し、後続フェーズの実装スコープを確定する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.1.1 | `docs/cc-2.1.99-2.1.110-impact.md` を新設し、各バージョンの変更点を行ごとに列挙、`.claude/rules/cc-update-policy.md` の A (実装あり) / B (書いただけ) / C (CC 自動継承) に分類する。同時に `docs/opus-4-7-impact.md` を新設し、Opus 4.7 の 8 つの変更 (literal instruction, xhigh, task budgets, tokenizer, vision 2576px, memory, /ultrareview, Auto Mode 拡大) を Harness の影響箇所 (agents/, skills/, guardrails, hooks.json, docs/) にマッピングする | (a) CC 2.1.99-110 の全主要項目が分類される (B ゼロ)、(b) Opus 4.7 の 8 項目全てに「影響あり/なし + 対応方針」が記載、(c) 44.2 以降のタスクが分類表からトレース可能 | - | cc:完了 [72ebe35] |

---

### Phase 44.2: CC 2.1.105 `PreCompact` hook + `monitors` manifest 統合 [P0]

Purpose: 長時間 Worker の安定性を支える CC 2.1.105 の新機構を Harness に取り込む

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.2.1 | `.claude-plugin/hooks.json` に `PreCompact` hook を追加し、`go/cmd/harness/pre_compact.go` (新設) で (a) 長時間 Worker 実行中は block (`{"decision":"block"}` または exit 2)、(b) reviewer/advisor セッションは許可、(c) Plans.md 未保存の状態では block してエスカレーション、の 3 判定を実装する。既存 PreToolUse / PostToolUse / Stop hook とは干渉しない | (a) `tests/test-pre-compact-hook.sh` (新設) で 3 判定が PASS、(b) `go test ./go/internal/hookhandler/...` PASS、(c) Breezing 長時間セッションで compaction が走らないことを手動確認 | 44.1.1 | cc:完了 [b7e4263] |
| 44.2.2 | `.claude-plugin/plugin.json` に top-level `monitors` manifest を追加し、session 起動時に (a) harness-mem 健全性、(b) advisor/reviewer 状態、(c) Plans.md-実装の drift、を auto-arm でストリーミング監視する。既存 `scripts/monitor-*.sh` を `monitors` manifest から呼び出せるように統一する | (a) `claude plugin validate` PASS、(b) `/skills` 呼び出しで monitor が auto-arm される、(c) `go/cmd/harness/monitor_status.go` (新設) で monitor 一覧が出力できる | 44.1.1 | cc:完了 [a1aae68] |

---

### Phase 44.3: Guardrails R01-R13 再検証 [P0]

Purpose: CC 2.1.110 の `PermissionRequest updatedInput` deny 再チェック / `PreToolUse additionalContext` 永続化 / v2.1.101 `permissions.deny` 上書き修正、v2.1.98 の Bash permission bypass fix 群に対応して、Harness の guardrails が意図通り動くことを再確認する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.3.1 | `go/internal/hookhandler/guardrails/` 配下の R01-R13 実装を CC 2.1.110 仕様に再適合させる。特に (a) `PermissionRequest` が `updatedInput` + `setMode` を返した場合にも `settings.json` の deny が再評価されるパス、(b) `PreToolUse` の `additionalContext` が tool 失敗後も破棄されないパス、(c) backslash-escape / compound command / env-var prefix の Bash bypass を閉じるパス、を回帰テスト化する。deny 一覧は `.claude-plugin/settings.json` の現行 14 項目を基準に維持 | (a) `go/internal/hookhandler/guardrails/*_test.go` に 3 新規シナリオが入り PASS、(b) `tests/test-guardrails-r01-r13.sh` (新設または既存拡張) PASS、(c) `.claude/rules/self-audit.md` の integrity マーカー更新確認 | 44.1.1 | cc:完了 [f921cd98] |

---

### Phase 44.4: Opus 4.7 literal prompt re-tune [P0]

Purpose: Opus 4.7 の literal instruction following に合わせて、Harness 全体のプロンプト・エージェント指示・スキル説明を曖昧表現ゼロに書き直す。**本 Phase は既存動作の回帰リスクが最大なので最優先**

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.4.1 | `agents/worker.md`, `agents/reviewer.md`, `agents/advisor.md`, `agents/scaffolder.md`, `agents/team-composition.md` の initialPrompt / description / 本文から曖昧表現 (「いい感じに」「必要に応じて」「適切に」「ちゃんと」「なるべく」等) を除去し、具体的な判断基準 (閾値・マーカー・ファイル名・コマンド名) に置換する。`.claude/rules/opus-4-7-prompt-audit.md` (新設) に audit checklist + before/after 例を記録 | (a) agents/*.md に対する grep で対象曖昧表現ヒット 0、(b) audit checklist 全項目 PASS、(c) Codex mirror (`agents-codex/`) も同期 | 44.1.1 | cc:完了 [1a0f9d0] |
| 44.4.2 | `skills/harness-work`, `skills/harness-review`, `skills/harness-plan`, `skills/harness-release`, `skills/harness-setup`, `skills/harness-loop`, `skills/breezing` の SKILL.md + references/ を同じ audit checklist で literal 化する。description / description-ja も含め、Opus 4.7 が正しく skill auto-load できるよう trigger phrase を明確化 | (a) audit checklist PASS、(b) `tests/validate-plugin.sh` PASS、(c) Codex/opencode mirror 同期 (`check-consistency.sh` PASS) | 44.4.1 | cc:完了 [856e9888] |

---

### Phase 44.5: `xhigh` effort 採用判定 + effort frontmatter 更新 [P0]

Purpose: Opus 4.7 の新 `xhigh` effort を Harness の skill/agent frontmatter に採用するか、採用するならどのスコープで使うかを決定する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.5.1 | Claude Code の `effort` frontmatter (v2.1.72 以降 `low/medium/high` 3段階) と Opus 4.7 API の `xhigh` の対応関係を検証する (CC が API に渡す実効値を確認)。対応が取れる場合は `skills/harness-review/SKILL.md`, `agents/reviewer.md`, `agents/advisor.md` の effort を `xhigh` に引き上げる。取れない場合は `docs/effort-level-policy.md` (新設) に「CC v2.1.X 対応まで `high` を維持」と rationale を書く。v2.1.94 以降 default effort が `medium→high` に上がっている点も反映 | (a) CC の effort と API effort の対応マトリクスが docs/effort-level-policy.md に明記、(b) 決定 (採用 or 見送り) に従い frontmatter 更新 or 維持、(c) `tests/validate-plugin.sh` PASS | 44.4.2 | cc:完了 [315de2b9] |

---

### Phase 44.6: `ENABLE_PROMPT_CACHING_1H` opt-in [P1]

Purpose: CC v2.1.108 の 1 時間 prompt cache を Harness の長時間セッション (Breezing / harness-loop) で使い、コスト削減を確実に取る

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.6.1 | `scripts/enable-1h-cache.sh` (新設) で `ENABLE_PROMPT_CACHING_1H=1` を env.local 経由で opt-in 可能にする。`skills/breezing/SKILL.md` と `skills/harness-loop/SKILL.md` に「長時間セッション開始前に推奨」の 1 行を追加。`docs/long-running-harness.md` に 1-hour vs 5-minute cache の選択基準 (セッション長 >30 分なら 1h) を追記 | (a) `tests/test-prompt-cache-1h.sh` (新設) で env 伝播を検証 PASS、(b) 手動で Breezing セッションを回し `/cost` でキャッシュヒット率の増加を確認、(c) docs 更新完了 | 44.2.2 | cc:完了 [a8f1c308] |

---

### Phase 44.7: CC 2.1.99-110 小機能の Harness 導入 [P1]

Purpose: `EnterWorktree path` / `/recap` / `/undo` / model-invoked built-in slash commands / skill mid-message fix など、単発の改善を Harness に取り込む

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.7.1 | (a) `EnterWorktree` を使う箇所 (scripts/run-worker-*.sh 等) で `path` 引数による既存 worktree 再入を利用、(b) `/recap` を `skills/session-memory` の推奨フローに追加、(c) `/undo` を `.claude/rules/commit-safety.md` (必要なら新設) に記録、(d) `disable-model-invocation: true` で定義済みの Harness skill の mid-message 呼び出しを smoke test、(e) model が built-in slash commands (`/init`, `/review`, `/security-review`) を Skill tool から呼ぶケースの Harness 側影響を docs に 1 段落追記 | (a)-(e) 全て実装 or 判断記録完了、(f) `tests/validate-plugin.sh` PASS、(g) smoke test PASS | 44.2.2 | cc:完了 [c7be2b5c] |

---

### Phase 44.8: `/ultrareview` と `/harness-review` の連携 [P1]

Purpose: Opus 4.7 付属の `/ultrareview` が Harness の `/harness-review` と機能重複するため、並立 / 委譲 / 統合のいずれかに方針を確定する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.8.1 | `/ultrareview` の挙動を実環境で検証 (single-turn dedicated review session)。`/harness-review` との差分 (Harness は Plan-Work-Review の Plans.md 連動 + Codex adversarial + sprint contract 検証あり) を `docs/ultrareview-policy.md` (新設) に表で整理。決定: (A) 並立維持 + `/harness-review` 内で `/ultrareview` を opt-in サブステップとして呼ぶ or (B) `/harness-review` を優先し `/ultrareview` を触らない、のどちらかを選ぶ。選んだ方針に従い `skills/harness-review/SKILL.md` を更新 | (a) 決定が `docs/ultrareview-policy.md` に書かれる、(b) `skills/harness-review/SKILL.md` が方針に従って更新される、(c) mirror 同期 | 44.4.2 | cc:完了 [e24f7f99] |

---

### Phase 44.9: Opus 4.7 Vision 2576px 対応 [P2]

Purpose: Opus 4.7 の 3 倍以上に増えた画像解像度対応を、`harness-review` の PDF/スクリーンショット系フローで活かす

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.9.1 | `skills/harness-review/references/` に vision 高解像度フロー (PDF pages, 設計図画像レビュー) を 1 枚追加。`docs/opus-4-7-vision-usage.md` (新設) に「2576px まで安全」「超える場合は事前リサイズ」のガイドを書く。実装は skill references 追加のみで、新 API 呼び出しコードは不要 | (a) references 追加、(b) docs 追加、(c) `tests/validate-plugin.sh` PASS | 44.4.2 | cc:完了 [5949ee7d] |

---

### Phase 44.10: Task Budgets (public beta) 調査メモ [P2]

Purpose: Task Budgets を将来採用するための判断材料を整理する (本 Phase では採用しない)

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.10.1 | `docs/task-budgets-research.md` (新設) に Task Budgets (public beta) の API 仕様、Harness 既存の `max_consults` / `/cost` との競合関係、採用するなら「どの skill で」「どの粒度で」を整理する。本 Phase では実装しない判断を rationale 付きで記録 | (a) docs 新設、(b) 採用判断の次サイクル回し | 44.5.1 | cc:完了 [95d1b39b] |

---

### Phase 44.11: Feature Table + docs 更新 [P0]

Purpose: v2.1.99-v2.1.110 と Opus 4.7 の全項目を Feature Table に記載し、A/B/C 分類を可視化する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.11.1 | `docs/CLAUDE-feature-table.md` に v2.1.99-v2.1.110 の全主要エントリを追加 (付加価値列に A: 実装あり / C: CC 自動継承 を明記、B はゼロ)。あわせて Opus 4.7 セクションを新設し、literal instruction / xhigh / task budgets / tokenizer / vision / memory / /ultrareview / Auto Mode 拡大の 8 項目を記述。`CLAUDE.md` の「Claude Code Feature Utilization」行を "CC v2.1.110+ + Opus 4.7 の機能を活用" に更新 | (a) 追加行ごとに A/C 分類あり (B ゼロ)、(b) `.claude/rules/cc-update-policy.md` のゲート PASS、(c) `tests/validate-plugin.sh` PASS | 44.2.2, 44.3.1, 44.4.2, 44.5.1 | cc:完了 [d96e94b7] |

---

### Phase 44.12: Smoke test + CHANGELOG + v4.2.0 リリース [P0]

Purpose: 上記全タスクの統合動作を確認し、v4.1.1 → v4.2.0 "Arcana" リリースを完成させる

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.12.1 | CC 2.1.110 + Opus 4.7 環境で Plan → Work → Review のエンドツーエンド smoke test を実施。対象: `/harness-plan create`, `/harness-work`, `/breezing`, `/harness-review`, `/harness-release`。各フローで (a) literal prompt 変更後の挙動確認、(b) PreCompact block 動作、(c) monitors auto-arm、(d) 1h cache 反映、(e) xhigh effort (採用した場合) の効果、(f) guardrails R01-R13 の regression ゼロ、を確認し `docs/smoke-test-v4.2.0.md` (新設) に記録 | (a) smoke test 全 PASS、(b) regression ゼロ、(c) smoke test 記録完了 | 44.2.1, 44.2.2, 44.3.1, 44.4.2, 44.5.1, 44.6.1, 44.7.1, 44.8.1, 44.11.1 | cc:完了 [b8e1e386] |
| 44.12.2 | `CHANGELOG.md` の `[Unreleased]` セクションに v4.2.0 "Arcana" エントリを書く (日本語・「今まで / 今後」形式、CC 統合部分は「CC のアプデ → Harness での活用」形式)。`scripts/sync-version.sh bump minor` で `VERSION` / `.claude-plugin/plugin.json` / `harness.toml` を 4.2.0 に同期。GitHub Release notes は `.claude/rules/github-release.md` の英語フォーマットで下書き作成 | (a) CHANGELOG lint PASS、(b) 3 点バージョン同期、(c) `tests/validate-plugin.sh` + `scripts/ci/check-consistency.sh` + `scripts/check-residue.sh` 全 PASS、(d) GitHub Release 下書き保存 | 44.12.1 | cc:TODO |

---

### Phase 44.13: Plugin agent expose 経路の確立 (Agent tool subagent_type 不在バグ) [P0]

Purpose: 2026-04-17 の loop テストで発覚: harness の `agents/worker.md` / `reviewer.md` / `scaffolder.md` / `advisor.md` が CC runtime の Agent tool subagent_type enum に登録されておらず、`/harness-loop` / `/breezing` / `/harness-work` の Step 5 (worker spawn) が起動不可。

**重要な事前確認結果**:
- `claude plugin validate .claude-plugin/plugin.json` は `agents: Invalid input` `monitors: Invalid input` を返す → 公式 schema は両 field 不採用
- 比較: codex plugin (marketplace install) は plugin.json に `agents` 宣言なしで `codex:codex-rescue` を expose → agents は **auto-discovery される**
- `claude plugin list` に `claude-code-harness` が出ない → 本リポジトリは marketplace install ではなく **project-local** で load されている
- skills は project-local でも expose されるが、agents は marketplace install 経由のみで expose される疑い

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 44.13.1 | claude-code-harness の agents が Agent tool subagent_type enum に出ない原因を特定する。具体的には: (a) `claude plugin install` 経由で local marketplace install を試して `claude-code-harness:worker` が enum に出るか確認、(b) marketplace install 不要で公式に agents を expose する方法（例: `.claude-plugin/agents/` 配下に置く / 特定の frontmatter フラグ）を Claude Code docs から特定、(c) 確定した方法を README または `docs/plugin-development.md` (新設) に手順化、(d) agents/team-composition.md は agent 定義ではないので mv または excludeルールを追加 | (a) Agent tool で `subagent_type="claude-code-harness:worker"` を含む 4 agent が dry-run spawn 可能、(b) 解決手順がドキュメント化されている、(c) team-composition.md の扱いが明確 | 44.1.1 | cc:完了 [2026-04-18 検証] — 別フォルダでの install テスト結果: cache v4.1.1 plugin.json は agents field 無しでも `/agents` に harness:* 5 つ (worker/reviewer/scaffolder/advisor/team-composition) が FQN で表示。agents は `agents/` ディレクトリから **auto-discovery される** ことが確定。source repo 内セッションで出ない現象は development mode 的な CC 挙動で、end user 環境には影響しない開発者 localized な UX 問題。memory: `plugin_agent_auto_discovery.md` / `source_repo_plugin_hide.md` |
| 44.13.2 | `skills/harness-loop/SKILL.md` の frontmatter に `user-invocable: true` を追加。`Skill` tool 経由で直接 invoke 可能にする (短い名前 `harness-loop` で動作することを 2026-04-17 のテストで確認済み) | (a) `Skill(skill="harness-loop", args="...")` が動く、(b) `/loop /harness-loop ...` の dynamic mode も引き続き動作、(c) `tests/validate-plugin.sh` PASS | - | cc:完了 [skills/harness-loop/SKILL.md L7] |
| 44.13.3 | 44.13.1 の解決手段が「marketplace install 必須」と判明した場合のフォールバック: skill (harness-loop, breezing, harness-work) の Step 5 worker spawn を `subagent_type="general-purpose"` + worker prompt embedding に書き換え可能か検討。trade-off: agent 専用文脈が失われるが、project-local dev でも動作するメリットあり。実施は 44.13.1 結果次第 | (a) 44.13.1 結果を踏まえた decision record を `.claude/memory/decisions.md` に記録、(b) 採用なら 3 skill ファイルを更新し test 通す、(c) 不採用なら理由を decision record に記録して close | 44.13.1 | cc:不要 [2026-04-18] — 44.13.1 で「marketplace install で agents は auto-discovery で正常 expose される」が確定。fallback 不要。end user は `/plugin install` するだけで全 agents が使える |

---

## Phase 45: Plugin Manifest 公式準拠 + sync デグレ修正 + dead config 整理 (Phase 44 follow-up)

作成日: 2026-04-18 / 改訂: 2026-04-18 (公式 docs 確認後の方針転換)
目的: Phase 44 リリース直前に発覚した複数問題を一括解決する。

**当初の Phase 45 方針 (廃案)**: `pluginJSON` struct に `Monitors` / `Agents` field を追加して strip を防ぐ。

**改訂後の方針 (本 Phase で実装)**: 公式 plugins-reference 確認の結果、以下が判明したため公式準拠路線に転換する。
- `monitors` の公式 SSOT は `monitors/monitors.json` ファイル (plugin.json への直接記述ではない)
- `agents` は `agents/` directory auto-discovery が公式手段 (plugin.json への明示宣言は redundant)
- harness の独自 field (`id` / `autoArm`) は公式 schema 外 → `name` / `command` / `description` / `when` のみ使用
- harness.toml の `[telemetry]` セクションと `Telemetry` Go struct はパース後どこからも読まれない dead config

### 背景

- Phase 44.2.2 で `monitors` manifest を plugin.json に追加 → 2e8a3bb1 で誤って削除 → ca1e6e5d で restore → 36b73367 で再喪失、という流れで複数回事故が起きている
- 根本原因の理解の更新: 公式 docs (https://code.claude.com/docs/en/plugins-reference#monitors) によれば monitors の公式 SSOT は `monitors/monitors.json` であり、plugin.json に直接書く形式は技術的にはサポートされるが推奨ではない。harness が plugin.json に書いている形式は (1) 公式スキーマ外の `id` / `autoArm` を含む、(2) `claude plugin validate` が "Invalid input" を返す、(3) `harness sync` の `pluginJSON` struct に field がないため strip されやすい、という三重の脆弱性を持つ
- Agents 同様 `agents/` ディレクトリ auto-discovery が公式 SSOT。plugin.json の `agents` field は redundant declaration

### 優先度マトリクス (改訂後)

| 優先度 | Phase | 内容 | タスク数 | 依存 |
|--------|-------|------|---------|------|
| **Required** | 45.1 | monitors を `monitors/monitors.json` に公式準拠で移行 + plugin.json から削除 + sync.go で生成保証 | 1 | なし |
| **Required** | 45.2 | plugin.json から `agents` 削除 (auto-discovery のみ) + sync.go の出力確定 | 1 | 45.1 |
| **Required** | 45.3 | sync idempotent regression テスト (sync N 回後も plugin.json が安定) | 1 | 45.1, 45.2 |
| **Recommended** | 45.4 | dead config 整理: `[telemetry]` セクション + `Telemetry` Go struct + `otel_endpoint`/`webhook_url` 削除 | 1 | なし |
| **Recommended** | 45.5 | agent frontmatter 公式制限 (`hooks` / `permissionMode` 等が plugin agent では silently ignored) の調査記録 | 1 | なし |

合計: **5 タスク** (Required 3 / Recommended 2)

### 完成基準 (Definition of Done)

| # | 基準 | 検証方法 | 必須/推奨 |
|---|------|---------|----------|
| 1 | `monitors/monitors.json` が公式 schema (name/command/description/when) に準拠して存在 | docs URL https://code.claude.com/docs/en/plugins-reference#monitors と照合 | 必須 |
| 2 | `.claude-plugin/plugin.json` から `monitors` と `agents` field が削除されている | `cat plugin.json` で目視 + `claude plugin validate` PASS | 必須 |
| 3 | `harness sync` を 3 回連続実行しても `plugin.json` の checksum が不変 | `tests/test-sync-idempotent.sh` PASS | 必須 |
| 4 | `harness.toml` から `[telemetry]` セクション削除、Go struct から `Telemetry` 削除、関連テスト調整 | `go test ./go/...` PASS | 推奨 |
| 5 | `docs/agent-frontmatter-policy.md` (新設) に公式制限と現行 worker.md/reviewer.md/advisor.md/scaffolder.md の影響範囲を記録 | docs ファイル存在 + 修正方針記載 | 推奨 |

---

### Phase 45.1: monitors を公式 `monitors/monitors.json` に移行 [P0]

Purpose: 公式 plugins-reference の SSOT 形式に従い、monitors manifest の strip 問題を恒久解消する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 45.1.1 | (a) `monitors/monitors.json` (新設) に公式 schema (`name` / `command` / `description` / `when`) で `harness-session-monitor` 1 entry を記述。`id` 廃止、`autoArm: ["plugin_enable", "skill_invoke"]` は `when: "always"` (default 動作と同じ) で明示。(b) `.claude-plugin/plugin.json` から `monitors` block を完全削除。(c) `go/cmd/harness/sync.go` の `pluginJSON` struct には Monitors field を **追加しない** (公式 SSOT が file 側になったため、plugin.json で manifest を持たない設計を選択)。(d) `harness sync` 実行後も plugin.json に monitors が再出現しないことを確認 | (a) `monitors/monitors.json` が公式 schema 準拠で存在、(b) plugin.json に monitors なし、(c) `bash bin/harness sync` 実行後 plugin.json が clean、(d) `tests/validate-plugin.sh` で `claude plugin validate` の monitors-related エラーゼロ | - | cc:完了 [d3e5c8c7] |

---

### Phase 45.2: plugin.json から agents 削除 (auto-discovery 一本化) [P0]

Purpose: 公式 docs 通り `agents/` ディレクトリ auto-discovery のみで agents を expose する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 45.2.1 | (a) `.claude-plugin/plugin.json` から `agents` field を削除。(b) `go/cmd/harness/sync.go` の `pluginJSON` struct に Agents field を **追加しない** (45.1.1 と同方針)。(c) auto-discovery で agents が `/agents` UI に表示されることをスモーク確認 (memory: `plugin_agent_auto_discovery.md` で marketplace install 経由動作確定済み) | (a) plugin.json に agents なし、(b) `harness sync` 後も agents が再出現しない、(c) `claude plugin validate` で agents-related エラーゼロ | 45.1.1 | cc:完了 [d3e5c8c7 — 45.1.1 と同 commit で副次的に達成、別 commit 不要] |

---

### Phase 45.3: idempotent sync の regression テスト [P0]

Purpose: sync を複数回実行しても plugin.json が安定 (公式 SSOT 外の field が混入しない) ことを保証する

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 45.3.1 | `tests/test-sync-idempotent.sh` (新設) で (a) plugin.json checksum 記録、(b) `bash bin/harness sync` を 3 回連続実行、(c) checksum が初回と一致、(d) `monitors` / `agents` field が plugin.json に存在しないことを assert。Go unit test `go/cmd/harness/sync_no_phantom_fields_test.go` (新設) で `pluginJSON` struct が monitors/agents を **emit しない** ことも検証。`tests/validate-plugin.sh` のセクション 9 か新セクションでこのテストを呼び出す | (a) shell test PASS、(b) Go unit test PASS、(c) `tests/validate-plugin.sh` PASS | 45.1.1, 45.2.1 | cc:完了 [ef0eee75] |

---

### Phase 45.4: dead config 整理 ([telemetry] セクション削除) [P1]

Purpose: harness.toml の `[telemetry]` セクションと `Telemetry` Go struct はパース後どこからも参照されていない dead config。誤って toml に書けば動くと誤解する利用者を防ぐ

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 45.4.1 | (a) `harness.toml` から `[telemetry]` セクション (otel_endpoint, webhook_url コメントアウト含む) を削除。(b) `go/pkg/config/toml.go` から `TelemetryConfig` struct と `Telemetry` field を削除。(c) `go/pkg/config/toml_test.go` から telemetry 関連 test ケースを削除または「NOT 設定」前提に書き換え。(d) `go/cmd/harness/init.go` の scaffold template から `[telemetry]` セクション例を削除。(e) `webhook_url` は env 変数 `HARNESS_WEBHOOK_URL` 経由のみで設定される旨を `docs/long-running-harness.md` または `CLAUDE.md` に 1 行追記 | (a)-(e) 完了、(f) `go test ./go/...` PASS、(g) `tests/validate-plugin.sh` PASS、(h) `bash bin/harness sync` 動作 | - | cc:TODO |

---

### Phase 45.5: agent frontmatter 公式制限の調査記録 [P1]

Purpose: 公式 plugins-reference によれば plugin agent では `hooks` / `mcpServers` / `permissionMode` は silently ignored される (security restriction)。harness の `worker.md` / `reviewer.md` / `advisor.md` / `scaffolder.md` がこれらを使用しているため、想定動作と実動作の乖離を docs に記録する。本 Phase では実装変更しない (調査と判断記録のみ)

| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 45.5.1 | `docs/agent-frontmatter-policy.md` (新設) に以下を含める: (a) 公式 plugins-reference の plugin agent 対応 frontmatter 一覧 (name/description/model/effort/maxTurns/tools/disallowedTools/skills/memory/background/isolation のみ)、(b) harness の各 agent (worker/reviewer/advisor/scaffolder) で使用している全 frontmatter field のリストと、公式対応状況の表、(c) silently ignored される `hooks` / `permissionMode` の影響範囲分析 (例: Worker の bypassPermissions は実は効いていない可能性、PreToolUse hook も plugin agent では発火しない可能性)、(d) memory `worker_worktree_share.md` の現象との関連可能性、(e) 修正案 (例: `hooks` を `hooks/hooks.json` に SubagentStart matcher で外出し / `permissionMode` の代替案検討)、(f) 本 Phase で実装しない rationale (リスク評価とリリース優先度) | (a) docs 新設、(b) 各 agent の frontmatter 監査表あり、(c) 修正案 3 つ以上、(d) 本 Phase 見送り判断の rationale 明記 | - | cc:TODO |

---
