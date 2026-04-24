<p align="center">
  <img src="docs/images/claude-harness-logo-with-text.png" alt="Claude Harness" width="400">
</p>

<p align="center">
  <strong>Plan. Work. Review. Ship.</strong><br>
  <em>Claude Code を規律ある開発パートナーに変える</em>
</p>

<p align="center">
  <a href="https://github.com/Chachamaru127/claude-code-harness/releases/latest"><img src="https://img.shields.io/github/v/release/Chachamaru127/claude-code-harness?display_name=tag&sort=semver" alt="Latest Release"></a>
  <a href="LICENSE.md"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
  <a href="docs/CLAUDE_CODE_COMPATIBILITY.md"><img src="https://img.shields.io/badge/Claude_Code-v2.1+-purple.svg" alt="Claude Code"></a>
  <img src="https://img.shields.io/badge/Skills-5_Verbs-orange.svg" alt="Skills">
  <img src="https://img.shields.io/badge/Core-Go_Native-00ADD8.svg" alt="Go Core">
  <img src="https://img.shields.io/badge/v4.2-Hokage-FF4500.svg" alt="Hokage">
</p>

<p align="center">
  <a href="README.md">English</a> | 日本語
</p>

<p align="center">
  <img src="docs/images/hokage/hokage-hero.jpg" alt="Hokage v4.0 — The Silent Blade" width="860">
</p>

---

## v4.2 アップデート — Claude Code 2.1.99-110 + Opus 4.7 完全追従

> **Hokage ライン継続。Claude Code 2.1.99-2.1.110 + Opus 4.7 の機能群に Harness を完璧に追従。plugin manifest を公式 `plugins-reference` 準拠に移行。**

Anthropic がリリースした Opus 4.7 は literal-instruction-following（書かれたとおりに実行する）セマンティクスを採用、Claude Code 2.1.105 では `PreCompact` フックと `monitors/monitors.json` マニフェストが追加されました。v4.2 はこれらに合わせて Harness を再調整します:

| 領域 | これまで (v4.1) | これから (v4.2) |
|---|---|---|
| **長時間 Worker** | タスク途中で勝手に context 圧縮されることがあった | `PreCompact` フックが、Worker 実行中または `Plans.md` が dirty な間は圧縮を block |
| **Plugin 検証** | `claude plugin validate` が `monitors`/`agents` ブロックを Invalid input で reject | 公式準拠: `monitors/monitors.json` + `agents/` auto-discovery |
| **Sync の事故** | `harness sync` が宣言済みブロックを黙ってストリップ (過去 4 回事故発生) | 二層ガード: shell 冪等性テスト + Go struct phantom field テスト |
| **長時間セッション** | デフォルト 5 分 prompt cache のみ | `bash scripts/enable-1h-cache.sh` で 1 時間 TTL に opt-in (CC v2.1.108) |
| **Reviewer/Advisor effort** | `medium` / `high` | `xhigh` (CC v2.1.111, Opus 4.7) — レビュー精度向上、他モデルでは `high` フォールバック |
| **Agent prompts** | Opus 4.6 の暗黙補完前提 | Opus 4.7 literal instruction following 対応に再チューニング — 閾値・スキーマ・コマンド名を明示 |
| **Guardrails (R01-R13)** | CC 2.1.98 仕様準拠 | CC 2.1.110 仕様に再適合 (`PermissionRequest updatedInput`、`PreToolUse additionalContext`、Bash bypass 閉鎖) — 17 件の回帰テスト追加 |

**体感で変わること:**
- 長時間タスクが自動 compaction で途中切断されなくなる
- `claude plugin validate` が `monitors` 追加以来初めてクリーンに通る
- `harness sync` で `monitors`/`agents` ブロックが勝手に消える事故がなくなる
- Opus 4.7 で動かすと Reviewer エージェントのフィードバックが鋭くなる (xhigh effort)

通常のフローで更新:
```
/plugin update claude-code-harness
```

詳細な Before/After は `CHANGELOG.md` の `[4.2.0]` エントリを参照。

---

## v4.0 "Hokage" — 何が変わったか

> **Go ネイティブエンジン。25倍高速なフック。Node.js 依存ゼロ。**

Claude が使うすべてのツール呼び出しは Harness のフックを通過します。v3 では毎回 40-60ms の bash + Node.js オーバーヘッドがかかっていました — 1セッションで数百回の呼び出しを考えると、積み重なる微妙な引っかかり。v4 はスタック全体を単一の Go バイナリに置き換えました:

| | v3（移行前） | v4 "Hokage"（移行後） |
|---|---|---|
| **PreToolUse** | 40-60ms | **10ms** |
| **SessionStart** | 500-800ms | **10-30ms** |
| **PostToolUse** | 20-30ms | **10ms** |
| **Node.js** | 必要 (18+) | **不要** |

**体感で変わること:**
- ツール呼び出し間の微妙な待ち時間が消える — Claude の応答がスムーズに感じる
- セットアップ時の `npm install` や Node.js バージョン問題がなくなる
- セッション起動が一瞬 (10-30ms、以前は約1秒)
- オプションの [harness-mem](https://github.com/Chachamaru127/harness-mem) 連携: 前回のセッションで何をしたか覚えている

プラグインを更新するだけ — 設定変更は不要:
```
/plugin update claude-code-harness
```

---

## なぜ Harness？

Claude Code は強力です。Harness はその力を、信頼しやすく、途中で崩れにくい開発フローへ変えます。

<p align="center">
  <img src="assets/readme-visuals-ja/generated/why-harness-pillars.svg" alt="Harness を入れると変わること: 計画の定着、実行時ガード、やり直せる検証" width="860">
</p>

5動詞スキルで流れを揃え、Go ネイティブガードレールエンジンで実行を守り、検証は必要なときに同じ手順でやり直せます。

## 人気の Claude Code ハーネスと比べると

ここで見たいのは、Claude Code が理論上どこまでできるかではなく、ハーネスを入れたあとに **標準の進め方がどう変わるか** — そして [harness-mem](https://github.com/Chachamaru127/harness-mem) と組み合わせたときに何が加わるかです。

<p align="center">
  <img src="assets/readme-visuals-ja/generated/harness-feature-matrix.svg" alt="Claude Harness + harness-mem vs Harness 単体 vs Superpowers vs cc-sdd — 10項目比較" width="860">
</p>

harness-mem と組み合わせれば、Claude Harness は計画・実装・レビュー・検証が崩れにくく、**セッション間の記憶も引き継がれる**唯一のセットアップです。

根拠とソース一覧: [docs/github-harness-plugin-benchmark.md](docs/github-harness-plugin-benchmark.md)

---

## 動作要件

- **Claude Code v2.1+** ([インストールガイド](https://docs.anthropic.com/claude-code))
  - v2.1.105+ 推奨 (PreCompact フック + monitors manifest)
  - v2.1.111+ で `xhigh` effort と Opus 4.7 サポート
- **Opus 4.7** (`claude-opus-4-7`) 推奨 — v4.2 のフル機能 (literal instruction following / vision 2576px / xhigh effort) を活かすため
- **Node.js 不要** (v4.0 Hokage は Go ネイティブエンジン)

---

## 誰のためのツール？

| あなたが | Harness でできること |
|----------|---------------------|
| **開発者** | 組み込み QA で高速に出荷 |
| **フリーランサー** | クライアントにレビューレポートを納品 |
| **インディーハッカー** | 壊さずに素早く動く |
| **VibeCoder** | 自然言語でアプリを構築 |
| **チームリード** | プロジェクト横断で標準を強制 |

---

## 30秒でインストール

```bash
# プロジェクトで Claude Code を起動
claude

# マーケットプレイスを追加してインストール
/plugin marketplace add Chachamaru127/claude-code-harness
/plugin install claude-code-harness@claude-code-harness-marketplace

# プロジェクトを初期化
/harness-setup
```

これだけ。`/harness-plan` から始めよう。

### 言語設定

配布時の既定言語は English です。日本語で使う場合は、明示的に opt-in します。

```yaml
i18n:
  language: ja
```

一時的に日本語セットアップで起動する場合は
`CLAUDE_CODE_HARNESS_LANG=ja claude` を使えます。英語版 README は
[README.md](README.md) です。

---

## 🪄 説明が長い？ならこれ: 検証前提の /work all

**読むのが面倒？** これだけ打てばいい:

```
/harness-work all
```

**計画承認後の最短導線はこれです。** 計画 → 並列実装 → レビュー → コミット。

<p align="center">
  <img src="assets/readme-visuals-ja/work-all-flow.svg" alt="/work all パイプライン" width="700">
</p>

> ⚠️ **実験的ワークフロー**: 計画を承認したら、Claude が完走します。実運用の前に [docs/evidence/work-all.md](docs/evidence/work-all.md) で成功系/失敗系の契約を確認してください。

---

## 5動詞ワークフロー

<p align="center">
  <img src="assets/readme-visuals-ja/generated/core-loop.svg" alt="Plan → Work → Review サイクル" width="560">
</p>

### 0. Setup（初期化）

```bash
/harness-setup
```

以後の作業が同じルールとコマンド面で走るように、プロジェクトファイルと初期設定を揃えます。

### 1. Plan（計画）

```bash
/harness-plan
```

> 「メールバリデーション付きのログインフォームが欲しい」

Harness が明確な受入条件付きの `Plans.md` を作成。

### 2. Work（実装）

```bash
/harness-work              # 並列数を自動検出
/harness-work --parallel 5 # 5ワーカーで同時実行
```

各ワーカーが実装、セルフレビュー、報告を行う。

### Advisor Strategy（相談役つき実行）

Harness は、少しずつ Advisor Strategy を取り入れています。
これは、**実行役がふだんは自走し、本当に難しい場面だけ相談役を呼ぶ**進め方です。

- 高リスク task では、最初に 1 回だけ相談することがある
- 同じ原因の失敗が続いたら、次の打ち手を相談する
- plateau 検知では、止める前に最後の 1 回だけ相談する
- 最終的な合否判定は、これまで通り Reviewer が持つ

最初の導入先は `harness-loop` です。
長時間実行でいちばん効果が出やすく、相談履歴も追いやすいためです。

詳しくは [docs/advisor-strategy.md](docs/advisor-strategy.md) を参照してください。

<p align="center">
  <img src="assets/readme-visuals-ja/parallel-workers.svg" alt="並列ワーカー" width="640">
</p>

### 3. Review（レビュー）

```bash
/harness-review
```

<p align="center">
  <img src="assets/readme-visuals-ja/review-perspectives.svg" alt="4視点レビュー" width="640">
</p>

| 視点 | 焦点 |
|------|------|
| Security | 脆弱性、インジェクション、認証 |
| Performance | ボトルネック、メモリ、スケーリング |
| Quality | パターン、命名、保守性 |
| Accessibility | WCAG準拠、スクリーンリーダー |

### 4. Release（リリース）

```bash
/harness-release
```

実装とレビューの結果を前提に、CHANGELOG、タグ、リリース用のハンドオフをまとめます。

---

## セーフティファースト

<p align="center">
  <img src="assets/readme-visuals-ja/generated/safety-guardrails.svg" alt="安全保護システム" width="640">
</p>

Harness v4 は **Go ネイティブガードレールエンジン**（`go/internal/guardrail/`）でコードベースを保護 — 13 の宣言的ルール（R01–R13）、単一バイナリで応答 10ms 以下:

| ルール | 保護対象 | アクション |
|--------|----------|------------|
| R01 | `sudo` コマンド | **拒否** |
| R02 | `.git/`, `.env`, シークレット | 書き込み**拒否** |
| R03 | `rm -rf /`, 破壊的パス | **拒否** |
| R04 | `git push --force` | **拒否** |
| R05–R09 | モード固有のガード | コンテキスト判定 |
| Post | `it.skip`, アサーション改ざん | **警告** |
| Perm | `git status`, `npm test` | **自動許可** |

<p align="center">
  <img src="assets/readme-visuals-ja/safety-shield.svg" alt="セーフティシールド" width="600">
</p>

---

## 5動詞スキル、設定不要

v4 で42スキルを **5つの動詞スキル**に統合。まずは動詞から入り、必要になったときだけ Breezing、Codex、2-Agent を足せます。

<table>
<tr>
<td align="center" width="20%"><h3>/plan</h3>アイデア → Plans.md</td>
<td align="center" width="20%"><h3>/work</h3>並列実装</td>
<td align="center" width="20%"><h3>/review</h3>4視点コードレビュー</td>
<td align="center" width="20%"><h3>/release</h3>タグ + GitHub Release</td>
<td align="center" width="20%"><h3>/setup</h3>プロジェクト初期化</td>
</tr>
</table>

<p align="center">
  <img src="assets/readme-visuals-ja/skills-ecosystem.svg" alt="スキルエコシステム" width="640">
</p>

### 主要コマンド

| コマンド | 機能 | 旧コマンド |
|----------|------|-----------|
| `/harness-plan` | アイデア → `Plans.md` | `/plan-with-agent`, `/planning` |
| `/harness-work` | 並列実装 | `/work`, `/breezing`, `/impl` |
| `/harness-work all` | 承認済み計画 → 実装 → レビュー → コミット | `/work all` |
| `/harness-review` | 4視点コードレビュー | `/harness-review`, `/verify` |
| `/harness-release` | CHANGELOG、タグ、GitHub Release | `/release-har`, `/handoff` |
| `/harness-setup` | プロジェクト初期化 | `/harness-init`, `/setup` |
| `/memory` | SSOT ファイルを管理 | — |
| `harness doctor --residue` | 削除済みコードへの残存参照を自動検出 | — |

---

## アーキテクチャ

```
claude-code-harness/
├── go/                # Go ネイティブガードレール + hookhandler エンジン
│   ├── cmd/harness/   #   CLI エントリーポイント (sync, doctor, validate)
│   ├── internal/      #   guardrail / hookhandler / state / lifecycle / breezing
│   └── pkg/           #   config / hookproto (公開 API)
├── bin/               # ビルド済み harness バイナリ (darwin-arm64/amd64, linux-amd64)
├── skills/            # 5動詞スキル + 拡張スキル (plan/execute/review/release/setup ほか)
├── agents/            # 3エージェント (worker / reviewer / scaffolder)
├── hooks/             # CC フック設定 (hooks.json)
├── scripts/           # 補助シェルスクリプト
└── templates/         # 生成テンプレート
```

---

## 高度な機能

### Breezing（Agent Teams）

自律エージェントチームでタスクリストを一気に完走:

```bash
/harness-work breezing all                    # 計画レビュー + 並列実装
/harness-work breezing --no-discuss all       # 計画レビューをスキップして即実装
```

<p align="center">
  <img src="assets/readme-visuals-ja/breezing-agents.svg" alt="Breezing エージェントチーム" width="640">
</p>

**Phase 0（計画議論）** がデフォルトで実行 — Planner がタスクの品質を分析し、Critic が計画を批判的にチェック。承認後にコーディング開始。8タスク以上は自動バッチ分割。

### Session Memory（harness-mem）

[harness-mem](https://github.com/Chachamaru127/harness-mem) が動いていれば、Harness がセッション中のイベント — 作業内容、使ったツール、セッションの終わり方 — を自動記録します。次回のセッション開始時にそのコンテキストが検索・取得できます。

- **harness-mem なし**: イベントは `.claude/state/memory-bridge-events.jsonl` にローカル記録（外部依存ゼロ）
- **harness-mem あり**: イベントがメモリサーバーにも送信され、セッション横断の検索・取得が可能に

設定不要 — Harness が harness-mem を自動検知します。

<details>
<summary><strong>Codex エンジン</strong></summary>

実装タスクを OpenAI Codex に並列委託。Codex が実装、セルフレビュー、報告。

```bash
/harness-work --codex API エンドポイントを5つ実装して
/harness-review --codex  # 4視点 + Codex セカンドオピニオン
```

> **セットアップ**: [Codex CLI](https://github.com/openai/codex) をインストールし API キーを設定。または `./scripts/setup-codex.sh --user` で Codex CLI 内から直接 Harness スキルを利用可能。

</details>

<details>
<summary><strong>2-Agent モード（Cursor連携）</strong></summary>

Cursor を PM として、Claude Code を実装者として使用。Plans.md が両者間で同期。

```bash
/harness-release handoff  # Cursor PM に報告
```

</details>

<details>
<summary><strong>コンテンツ生成（スライド & 動画）</strong></summary>

```bash
/generate-slide   # 3パターン、品質スコアリング、自動エクスポート
/generate-video   # JSON Schema 駆動パイプライン + Remotion
```

> **依存**: スライドは `GOOGLE_AI_API_KEY` が必要。動画は [Remotion](https://www.remotion.dev/) + ffmpeg が必要。

</details>

---

## なぜ skill pack だけではなく Harness か？

skill pack はプロンプトを教えられます。Harness は実行時のふるまいまで強制します。

- **ガードレールエンジン** が破壊的な書き込み、シークレット露出、force push 系を実行経路で止めます。
- **Hooks と review フロー** が、実際に repo を触るツールの近くで品質確認を行います。
- **検証スクリプトと evidence pack** が、docs・配布物・`/harness-work all` の主張を再実行可能にします。

---

## トラブルシューティング

| 問題 | 解決策 |
|------|--------|
| コマンドが見つからない | まず `/harness-setup` を実行 |
| Windows で `harness-*` コマンドが出ない | プラグインを更新または再インストールしてください。公開 command skill は実ディレクトリで配布されるようになり、`core.symlinks=false` でも隠れなくなります。 |
| プラグインが読み込まれない | キャッシュをクリア: `rm -rf ~/.claude/plugins/cache/claude-code-harness-marketplace/` して再起動 |
| フックが動作しない | `harness doctor` を実行して診断を確認 |
| 移行後に古い参照が残っている | `bin/harness doctor --residue` で削除済みコードへの参照を自動検出 |

詳しいヘルプは [Issue を作成](https://github.com/Chachamaru127/claude-code-harness/issues)してください。

---

## アンインストール

```bash
/plugin uninstall claude-code-harness
```

プロジェクトファイル（Plans.md、SSOT ファイル）はそのまま残ります。

---

## Claude Code 機能ハイライト

Harness は最新の Claude Code 機能を自動的に活用します。体感で変わる部分:

| 何ができるか | 仕組み |
|-------------|--------|
| **並列安全書き込み** | Worktree 分離で複数ワーカーが同一ファイルを編集可能 |
| **スマート effort 調整** | 複雑なタスクに ultrathink モードを自動適用 |
| **自動エスカレーション** | ツール連続失敗 3 回でリカバリ発動 |
| **LLM 品質ガード** | Agent hook が全編集にセキュリティ・品質チェックを実行 |
| **チーム監視** | Breezing がワーカーの完了・アイドルを自動検知 |
| **モデル柔軟性** | `modelOverrides` で任意プロバイダ（Bedrock, Vertex 等）に対応 |

全技術一覧 (19 機能): [docs/CLAUDE-feature-table.md](docs/CLAUDE-feature-table.md)

---

## ドキュメント

| リソース | 説明 |
|----------|------|
| [Changelog](CHANGELOG.md) | バージョン履歴 |
| [Claude Code 互換性](docs/CLAUDE_CODE_COMPATIBILITY.md) | 動作要件 |
| [Distribution Scope](docs/distribution-scope.md) | 配布対象 / 互換維持 / 開発専用の境界 |
| [Work All Evidence Pack](docs/evidence/work-all.md) | 成功系/失敗系の検証契約 |
| [Cursor 連携](docs/CURSOR_INTEGRATION.md) | 2-Agent セットアップ |
| [Benchmark Rubric](docs/benchmark-rubric.md) | static evidence と executed evidence の採点軸 |
| [Positioning Notes](docs/positioning-notes.md) | 公開向けの差別化メモ |
| [Content Layout](docs/content-layout.md) | docs と out の使い分けルール |

---

## コントリビュート

Issue と PR を歓迎します。[CONTRIBUTING.md](CONTRIBUTING.md) を参照。

---

## 謝辞

- [AI Masao](https://note.com/masa_wunder) — 階層的スキル設計
- [Beagle](https://github.com/beagleworks) — テスト改ざん防止パターン

---

## ライセンス

**MIT License** — 自由に使用、改変、商用利用可能。

[English](LICENSE.md) | [日本語](LICENSE.ja.md)
