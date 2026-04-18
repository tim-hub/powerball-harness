# Claude Code 2.1.99 → 2.1.110 — Harness 影響分類

> **目的**: Claude Code 2.1.99 から 2.1.110 までの全主要変更点を、
> `.claude/rules/cc-update-policy.md` の 3 カテゴリ (A / B / C) に分類し、
> Phase 44.2 以降の実装タスクへトレースできる形にまとめる。
>
> **分類ルール** (cc-update-policy.md より):
> - **A: 実装あり** — Harness 側に具体的変更が必要 (新しい hook / scripts / skills / agents / docs)
> - **B: 書いただけ** — Feature Table のみ変更。**禁止**。本表にあってはならない
> - **C: CC 自動継承** — CC 本体修正のみ。Harness 変更不要。Feature Table に「CC 自動継承」明記
>
> **前提**: 現 Harness は v4.1.1、Feature Table は v2.1.98 (Monitor ツール) まで対応済み。
> 本表の A 項目は Phase 44.2-44.7 で実装、C 項目は Phase 44.11 の Feature Table 追記のみ。

---

## 1. カテゴリ A: 実装あり（Harness 側変更が必要）

Phase 44 で Harness 側に具体的変更を伴う項目。各行に対応フェーズを明記する。

| ver | 変更点 | 影響箇所 | 対応 Phase |
|-----|-------|---------|-----------|
| 2.1.101 | Settings resilience: 不明な hook event 名が混入しても `settings.json` 全体が無視されなくなった | `.claude-plugin/settings.json` / `hooks.json` の新規 hook 追加時の安全性向上 | 44.2 (PreCompact 追加時の安全装置として利用) |
| 2.1.101 | `permissions.deny` が `PreToolUse` hook `permissionDecision: "ask"` を上書きするようになった | Harness guardrails (R01-R13) の deny 連鎖 | **44.3** (R01-R13 再検証) |
| 2.1.101 | Plugin 複数検出時の重複 `name:` frontmatter により slash command が別 plugin に解決されるバグ修正 | 既存の `skills/**/SKILL.md` frontmatter が一意名になっているかの監査 | **44.4.2** (skill literal 化時に name 一意性も検証) |
| 2.1.101 | Skill が `context: fork` と `agent` frontmatter を honor しない既存バグ修正 | Harness skill で `context: fork` を使っているスキル (canai-docs 等) の再検証 | **44.7.1** (小機能統合) |
| 2.1.101 | Subagent が動的注入 MCP server を継承しないバグ修正 | Breezing での動的 MCP、`harness-mem` 継承 | **44.7.1** |
| 2.1.101 | Sub-agent が isolated worktree 内の自身のファイルに Read/Edit できないバグ修正 | `isolation: worktree` の Worker / Advisor | **44.7.1** (動作確認 + smoke) |
| 2.1.105 | **新 hook: `PreCompact`** — `{"decision":"block"}` / exit 2 で compaction を停止可能 | 長時間 Worker の意図せぬ compaction 中断を防ぐ | **44.2.1** (Go 実装 + hooks.json 登録) |
| 2.1.105 | **Plugin manifest: `monitors` 新規 top-level key** — session 起動 / skill invoke で background monitor が auto-arm | Harness の mem 健全性 / drift 監視 / advisor 状態の常駐化 | **44.2.2** (plugin.json 追加) |
| 2.1.105 | `EnterWorktree` に `path` parameter 追加、既存 worktree への再入可能 | `scripts/run-worker-*.sh` 等の worktree 再利用 | **44.7.1** |
| 2.1.105 | `/proactive` alias for `/loop` | harness-loop のエイリアス方針 | **44.7.1** (docs 追記) |
| 2.1.108 | **`ENABLE_PROMPT_CACHING_1H` env var** — 1 時間 prompt cache TTL | Breezing / harness-loop の長時間セッションコスト削減 | **44.6.1** (opt-in スクリプト + docs) |
| 2.1.108 | `/recap` / `/undo` alias for `/rewind` | session-memory / commit safety | **44.7.1** |
| 2.1.108 | Model が built-in slash commands (`/init`, `/review`, `/security-review`) を Skill tool から呼べる | Harness の `/harness-review` との機能重複確認 | **44.7.1** / **44.8.1** |
| 2.1.110 | `/tui` command + `tui` setting (fullscreen rendering) | 運用ガイド更新 (docs) | **44.7.1** (docs 追記のみ、Harness 動作変更なし) |
| 2.1.110 | **Push notification tool** (Remote Control + "Push when Claude decides" 設定時) | `harness-loop` の長時間実行完了通知で利用可能 | **44.7.1** (docs 追記、将来採用の可能性を記録) |
| 2.1.110 | **`PermissionRequest` hooks が `updatedInput` を返した場合、`permissions.deny` ルールが再チェックされる** | guardrails R01-R13 の deny 連鎖の整合性確認 | **44.3.1** (再検証必須) |
| 2.1.110 | `setMode:'bypassPermissions'` が `disableBypassPermissionsMode` を尊重 | Harness の bypass 方針維持 | **44.3.1** (docs 追記) |
| 2.1.110 | **`PreToolUse` hook `additionalContext` が tool 呼び出し失敗時も破棄されないように修正** | guardrails の deny 理由注入が失敗後も残る | **44.3.1** (回帰テスト追加) |
| 2.1.110 | Skills with `disable-model-invocation: true` が `/<skill>` mid-message 呼び出しで動くよう修正 | Harness の `/harness-work`, `/harness-review` 等で発生していた潜在バグの解消 | **44.7.1** (smoke test) |

**A 項目数**: **19 項目**
**実装割当**: 44.2 (2), 44.3 (3), 44.4.2 (1), 44.6.1 (1), 44.7.1 (10), 44.8.1 (1) + 44.3.1 内 3 件集約

---

## 2. カテゴリ C: CC 自動継承（Harness 変更不要）

Feature Table への追加のみ。Harness の実装変更は不要だが、利用ガイドや期待値は更新する。

| ver | 変更点 | Harness 側で享受される効果 |
|-----|-------|-------|
| 2.1.101 | メモリリーク修正 — 長セッションの virtual scroller で historical message-list が数十個保持されるバグ | 長時間 Breezing の RSS 安定化 |
| 2.1.101 | `--resume` / `--continue` の dead-end branch アンカーによる大セッション context loss 修正 | harness-loop wake-up 時の resume 信頼性向上 |
| 2.1.101 | 5 分ハードコード timeout を `API_TIMEOUT_MS` に従わせる修正 (ローカル LLM / extended thinking) | Opus 4.7 xhigh で長くなる thinking の安全性 |
| 2.1.101 | Bedrock SigV4 認証失敗修正 | Bedrock 経由で使うユーザーへの透過的改善 |
| 2.1.101 | Grep tool ENOENT → system `rg` フォールバック | 全 skill の Grep 信頼性 |
| 2.1.101 | `/btw` が毎回会話全体を disk 書き込みするバグ修正 | コンテキストコスト削減 |
| 2.1.101 | `/plugin update` ENAMETOOLONG 修正 | `/harness-setup` での plugin 更新安定化 |
| 2.1.101 | Directory-source plugins の stale cache 修正 | Harness dev の再読込安定化 |
| 2.1.101 | Custom keybindings が Bedrock / Vertex で読み込まれない修正 | マルチプロバイダ環境のキーバインド |
| 2.1.101 | 命令注入脆弱性修正: POSIX `which` fallback (LSP binary detection) | セキュリティ自動継承 |
| 2.1.105 | Image がキュー済みメッセージで drop されるバグ修正 | マルチモーダル投入の安定化 |
| 2.1.105 | Leading whitespace trim で ASCII art / indented diagram が壊れるバグ修正 | 設計図・表出力の信頼性 |
| 2.1.105 | `alt+enter` / `Ctrl+J` newline 挿入修正 | 編集体験 |
| 2.1.105 | One-shot scheduled task の再発火修正 (file watcher 後処理漏れ) | scheduled 運用の信頼性 |
| 2.1.105 | Team/Enterprise inbound channel notification 消失修正 | CC のマルチプレイヤー機能 |
| 2.1.105 | `/skills` menu scroll 修正 | UI |
| 2.1.107 | Extended-thinking indicator の表示改善 (hint を早く出す) | Opus 4.7 xhigh 時の UX |
| 2.1.108 | `/compact` が大会話で "context exceeded" で失敗する修正 | 長時間セッションの信頼性 |
| 2.1.108 | DISABLE_TELEMETRY 利用者が 1h cache を受けられない修正 | 44.6.1 opt-in の前提として必要 |
| 2.1.108 | Agent tool auto mode での safety classifier transcript overflow 時の permission prompt 修正 | Auto Mode 採用時の信頼性 |
| 2.1.108 | Bash tool が `CLAUDE_ENV_FILE` で末尾 `#` コメント行があると出力しないバグ修正 | Bash 実行の安定性 |
| 2.1.108 | `claude --resume <session-id>` で `/rename` のカスタム名/色が失われるバグ修正 | セッション管理 |
| 2.1.108 | Policy-managed plugins が初回 install と別 project から実行時に auto-update しないバグ修正 | Enterprise/Teams 配布 |
| 2.1.108 | `language` 設定時に diacritical marks (アクセント等) が drop されるバグ修正 | i18n |
| 2.1.109 | Extended-thinking indicator rotating progress hint | UX |
| 2.1.110 | MCP tool calls が SSE/HTTP server の接続 drop 中に無限 hang するバグ修正 | MCP 経由の信頼性 (harness-mem 等) |
| 2.1.110 | Non-streaming fallback retries の multi-minute hang 修正 | 長時間タスクの UX |
| 2.1.110 | Session cleanup が subagent transcripts を含む完全削除されるよう修正 | disk 節約 |
| 2.1.110 | `/skills` menu scroll 修正 (fullscreen) | UI |
| 2.1.110 | Remote Control session re-login prompt 修正 (stale session) | Remote Control UX |

**C 項目数**: **30 項目**

---

## 3. カテゴリ B: 書いただけ（禁止）

**空**。本ドキュメントの全項目は A または C に分類されている。`cc-update-policy.md` の「カテゴリ B 検出時は PR をブロック」ルールに従い、B は一切含まない。

---

## 4. Phase 44 実装トレース表

Phase 44.2 以降のタスクが本分類表のどの A 項目に対応するかを逆引きできるようにする。

| Phase | A 項目 (上表参照) |
|-------|---|
| 44.2.1 (PreCompact hook) | 2.1.105: `PreCompact` hook |
| 44.2.2 (monitors manifest) | 2.1.105: `monitors` manifest key |
| 44.3.1 (guardrails R01-R13 再検証) | 2.1.101: `permissions.deny` が PreToolUse ask 上書き / 2.1.110: `updatedInput` 再チェック / 2.1.110: `additionalContext` persist / 2.1.110: `setMode:'bypassPermissions'` + `disableBypassPermissionsMode` |
| 44.4.2 (skill literal 化) | 2.1.101: duplicate `name:` frontmatter バグの波及確認 |
| 44.6.1 (1h prompt cache opt-in) | 2.1.108: `ENABLE_PROMPT_CACHING_1H` |
| 44.7.1 (小機能統合) | 2.1.101: `context: fork` + agent / subagent MCP 継承 / worktree Read/Edit / 2.1.105: `EnterWorktree path` / `/proactive` / 2.1.108: `/recap` / `/undo` / built-in slash via Skill tool / 2.1.110: `/tui` / Push notification / `disable-model-invocation` mid-message fix |
| 44.8.1 (/ultrareview 連携) | 2.1.108: built-in slash 呼び出しの一環で `/ultrareview` を Skill tool から呼ぶ検討 |
| 44.11.1 (Feature Table 更新) | 上記 A 19 項目 + C 30 項目すべて |

---

## 5. メモ: セキュリティ関連の自動継承（特筆）

2.1.97 → 2.1.98 で Bash 権限 bypass (backslash-escape flag / compound command / env-var prefix / `/dev/tcp` redirect) を全て塞ぎ、2.1.101 で POSIX `which` fallback の command injection も修正。これらは全て C 継承だが、Harness の guardrails R01-R13 (Bash 系 deny) の前提条件が改善されているため、44.3.1 で「前提が変わっていないか」を再確認する価値がある。
