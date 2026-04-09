# Hokage-Harness スピンオフ設計レポート

> Generated: 2026-04-09
> Status: Draft — ブラッシュアップ中

---

## 1. エグゼクティブサマリー

claude-code-harness の Go ガードレールエンジン（13,906行 + テスト6,836行）を
**Hokage-Harness** として独立リポジトリにスピンオフする設計レポート。

**結論: 技術的に実現可能。Go コードは親リポの TS/shell に対するランタイム依存がゼロ。**

---

## 2. 抽出可能なコンポーネント

| コンポーネント | パッケージ | 行数 | 抽出難度 |
|-------------|----------|------|---------|
| ガードレールエンジン | internal/guardrail/ | 768 | そのまま |
| SQLite状態管理 | internal/state/ | 1,753 | そのまま |
| セッション管理 | internal/session/ | 1,432 | そのまま |
| イベントハンドラ | internal/event/ | 1,449 | そのまま |
| 並列オーケストレーション | internal/breezing/ | 1,268 | そのまま |
| CI統合 | internal/ci/ | 907 | そのまま |
| ライフサイクル状態機械 | internal/lifecycle/ | 1,257 | そのまま |
| フックプロトコル | pkg/hookproto/ | 113 | そのまま |
| 設定パーサ | pkg/config/ | 448 | そのまま |
| CLIコマンド | cmd/harness/ | 1,607 | 微修正 |
| SPEC + DESIGN | docs/ | 1,177 | 書き換え |

**合計: ~13,900行がそのまま移植可能**

---

## 3. 再設計するもの（持っていかない）

| コンポーネント | 現ハーネス | Hokage での方針 |
|-------------|----------|---------------|
| 32スキル (SKILL.md) | Markdown | 厳選して新規作成（5-10個） |
| 4エージェント | Markdown | 再設計（Hokageコンセプト） |
| 40+フック定義 | hooks.json | Go が自動生成（harness sync） |
| shell スクリプト群 | bash | 不要 — Go バイナリが全て担う |
| TS ガードレール | core/ | 不要 — Go版が完全置換 |
| mirror (codex/opencode/) | コピー | 不要 — SSOT一本化 |

---

## 4. 推奨リポジトリ構造

```
hokage-harness/
├── .claude-plugin/
│   └── plugin.json              # ← harness sync が自動生成
│
├── cmd/harness/                 # Go エントリポイント
│   ├── main.go                  # サブコマンドルーター
│   ├── init.go                  # harness init
│   ├── sync.go                  # harness sync (plugin.json等を生成)
│   ├── validate.go              # SKILL.md/agent バリデーション
│   └── doctor.go                # ヘルスチェック
│
├── internal/                    # プライベート実装
│   ├── guardrail/               # R01-R13 ガードレールエンジン
│   ├── state/                   # SQLite 状態管理
│   ├── session/                 # セッションライフサイクル
│   ├── event/                   # フックイベントハンドラ
│   ├── breezing/                # 並列オーケストレーション
│   ├── ci/                      # CI 統合
│   ├── lifecycle/               # 状態機械 + リカバリ
│   └── hook/                    # プロトコルコーデック
│
├── pkg/                         # 公開 API
│   ├── hookproto/               # フックプロトコル型定義
│   └── config/                  # harness.toml パーサ
│
├── skills/                      # Hokage 専用スキル (厳選)
│   ├── plan/
│   │   └── SKILL.md
│   ├── execute/
│   │   └── SKILL.md
│   ├── review/
│   │   └── SKILL.md
│   └── ...
│
├── agents/                      # Hokage 専用エージェント
│   ├── worker.md
│   ├── reviewer.md
│   └── scaffolder.md
│
├── bin/                         # ビルド成果物 (CC が自動でPATHに追加)
│   ├── harness-darwin-arm64
│   ├── harness-darwin-amd64
│   └── harness-linux-amd64
│
├── output-styles/               # カスタム出力スタイル
│
├── harness.toml                 # SSOT 設定ファイル
├── go.mod                       # ルートに配置（go install 対応）
├── go.sum
├── Makefile
├── VERSION
├── CHANGELOG.md
├── LICENSE
└── README.md
```

---

## 5. 現行ハーネスとの差別化

| 観点 | claude-code-harness | Hokage-Harness |
|------|---------------------|----------------|
| ランタイム | bash + Node.js + Go 混在 | **Go オンリー** — shell 0本 |
| フック応答 | 40-60ms (bash→node) | **1-3ms** (直接バイナリ) |
| 状態管理 | ファイル + SQLite | **SQLite 一本化** (pure-Go) |
| スキル数 | 32個 (汎用) | **5-10個** (コアに集中) |
| エージェント | 4体 | **3体** (Worker/Reviewer/Scaffolder) |
| 配布 | プラグインディレクトリ | **go install + プラグイン** 両対応 |
| CGO | なし | なし — クロスコンパイル容易 |
| TS依存 | core/ に残存 | **完全排除** |

---

## 6. 公式プラグイン仕様との適合性

| 仕様 | 活用方法 |
|------|---------|
| `bin/` 自動PATH追加 | Go バイナリを bin/ に置くだけで Bash ツールから使える |
| `${CLAUDE_PLUGIN_ROOT}` | フック定義で `${CLAUDE_PLUGIN_ROOT}/bin/harness hook pretool` |
| `${CLAUDE_PLUGIN_DATA}` | SQLite DB を永続化ディレクトリに配置 |
| バイナリフック | command タイプで Go バイナリ直接呼び出し (stdin JSON → stdout JSON) |
| `userConfig` | ユーザーごとの設定をプラグイン有効化時に取得 |
| スキル2%予算 | description 250文字以内、詳細は references/ に分離 |

---

## 7. 類似プロジェクト比較

| プロジェクト | 特徴 | Hokage との差 |
|------------|------|-------------|
| Claudey | Go バイナリでフック処理、セッション管理 | Hokage はガードレール + 状態機械が充実 |
| claude-tab-fix | Go でインデント修正 | 単機能 |
| rulebricks/guardrails | API ベースの allow/deny | Hokage はローカル完結で高速 |
| trailofbits/config | セキュリティ設定テンプレート | ランタイムなし |

---

## 8. 移行作業見積もり

| タスク | 工数 | 説明 |
|--------|------|------|
| go.mod リネーム + import 更新 | 1h | github.com/owner/hokage-harness |
| go/ → ルート昇格 | 1h | ディレクトリ再配置 |
| Makefile 更新 | 30m | パス参照修正 |
| VERSION / sync-version.sh | 1h | 独立バージョン管理 |
| harness.toml テンプレート | 2h | Hokage 用デフォルト設定 |
| plugin.json 生成ロジック | 1h | メタデータ更新 |
| コアスキル新規作成 (5本) | 4h | Plan/Execute/Review/Setup/Release |
| エージェント定義 (3本) | 2h | Worker/Reviewer/Scaffolder |
| SPEC.md / DESIGN.md 書き換え | 2h | Phase番号除去、独立ドキュメント化 |
| README.md | 2h | QuickStart、アーキテクチャ図 |
| GitHub Actions CI | 2h | テスト + クロスビルド + リリース |
| E2E テスト移植 | 2h | test-e2e.sh 適応 |
| **合計** | **~20h** | |

---

## 9. リスクと対策

| リスク | 深刻度 | 対策 |
|--------|--------|------|
| バイナリサイズ (~7MB × 3 = 21MB) | 中 | GitHub Releases 配布、go install 推奨 |
| SQLite の PLUGIN_DATA 配置 | 低 | harness init で自動設定 |
| スキル不足 (32→5-10個) | 低 | コアに集中、拡張はユーザー追加 |
| ブランド混乱 | 中 | README で「スピンオフ」明記 |
| 二重メンテナンス | 高 | 元リポ go/ を凍結し Hokage を正統後継とする |

---

## 10. 元リポとの関係（3つの選択肢）

| 選択肢 | 説明 | 推奨度 |
|--------|------|--------|
| A. submodule 参照 | 元リポから go/ 削除、Hokage を submodule | △ 複雑 |
| B. go/ 凍結 | 元リポは TS 版維持、Hokage が Go の正統後継 | ◎ 推奨 |
| C. 完全独立 | 元リポとの関係を切る | ○ シンプル |

---

## 11. Codex 互換性 ✅ 調査完了

**結論: Codex CLI v0.116+ でも使える。コーデック層の薄い適応のみで両対応可能。**

### 11.1 Codex フックシステム（v0.116+, 2026年3月〜）

Codex は Claude Code と同じ「stdin JSON → stdout JSON + exit code」プロトコルで外部バイナリを呼び出せる。

| Codex フックイベント | CC 対応イベント | 対応状況 |
|---|---|---|
| `PreToolUse` | `PreToolUse` | Bash ツールのみ |
| `PostToolUse` | `PostToolUse` | Bash ツールのみ |
| `SessionStart` | `SessionStart` | 対応 |
| `UserPromptSubmit` | `UserPromptSubmit` | 対応 |
| `Stop` | `Stop` | 対応 |

### 11.2 プラットフォーム互換マトリクス

| コンポーネント | CC | Codex | 互換度 | 作業量 |
|-------------|-----|-------|--------|-------|
| Go バイナリフック | `type: "command"` | `type: "command"` | **完全** | ゼロ |
| stdin JSON | hookproto | ほぼ同一 | **高** | lenient parsing |
| PreToolUse 出力 | `permissionDecision` | exit code 2 or JSON | **中** | codec 分岐 |
| SKILL.md | YAML + MD | YAML + MD | **完全** | ゼロ |
| プラグインマニフェスト | `.claude-plugin/` | `.codex-plugin/` | **類似** | 2ファイル生成 |
| ガードレール | hooks.json | hooks.json + rules/ | **中** | ルール生成追加 |
| 設定 | settings.json | config.toml + rules | **低** | 変換ロジック |

### 11.3 アーキテクチャ: デュアルコーデック方式

```
┌─────────────────────────────────────────────┐
│          Hokage-Harness (Go Binary)          │
│                                              │
│  ┌──────────┐  ┌────────────┐  ┌─────────┐ │
│  │ Guardrail│  │   State    │  │ Session │ │
│  │  Engine  │  │  Machine   │  │ Manager │ │
│  │ (共通)   │  │  (共通)    │  │ (共通)  │ │
│  └────┬─────┘  └─────┬──────┘  └────┬────┘ │
│       └──────────┬───┘──────────────┘       │
│            ┌─────┴─────┐                     │
│            │  Codec層   │ ← ここだけ分岐      │
│            ├───────────┤                     │
│            │ CC  │ CDX │                     │
│            └──┬──┴──┬──┘                     │
└───────────────┼─────┼────────────────────────┘
          stdout│     │stdout
     ┌──────────┘     └──────────┐
     ▼                           ▼
 Claude Code                   Codex
```

自動検出: `CLAUDE_PLUGIN_ROOT` → CC / `CODEX_HOME` → Codex / fallback → generic

### 11.4 harness sync デュアル出力

```
harness sync
  ├── .claude-plugin/plugin.json    # CC 用マニフェスト
  ├── .claude-plugin/hooks.json     # CC 用フック定義
  ├── .claude-plugin/settings.json  # CC 用パーミッション
  ├── .codex-plugin/plugin.json     # Codex 用マニフェスト
  ├── .codex/hooks.json             # Codex 用フック定義
  └── .codex/rules/harness.rules    # Codex ガードレール (prefix_rule DSL)
```

### 11.5 変更が必要なファイル

| ファイル | 変更内容 | 追加行数 |
|---------|---------|---------|
| `internal/hook/codec.go` | CC/Codex デュアル出力 | +40行 |
| `cmd/harness/sync.go` | .codex-plugin/ 生成 | +100行 |
| `pkg/hookproto/types.go` | Codex フィールド追加 | +10行 |
| **ガードレールエンジン本体** | **変更不要** | **0行** |

### 11.6 SKILL.md のクロスプラットフォーム互換

SKILL.md (YAML frontmatter + Markdown) は以下で動作確認済み:
- Claude Code
- Codex CLI
- Gemini CLI
- Cursor

Hokage のスキルを1回書けば、全プラットフォームで利用可能。

### 11.7 参考URL

- [Codex Hooks](https://developers.openai.com/codex/hooks)
- [Codex Skills](https://developers.openai.com/codex/skills)
- [Codex Plugin Build](https://developers.openai.com/codex/plugins/build)
- [codex-plugin-cc (公式CC連携)](https://github.com/openai/codex-plugin-cc)
- [codex-hooks (CC→Codexブリッジ)](https://github.com/hatayama/codex-hooks)

---

## 12. 次のステップ

1. [ ] リポジトリ構造の確定（セクション4）
2. [ ] go.mod リネーム + ルート昇格
3. [ ] デュアルコーデック実装（codec.go）
4. [ ] harness sync のデュアル出力
5. [ ] コアスキル5本の新規作成
6. [ ] エージェント定義3本
7. [ ] GitHub Actions CI 構築
8. [ ] README.md + QuickStart
9. [ ] 初回リリース (v0.1.0)

---

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-09 | 初版作成（4エージェント並列調査結果を統合） |
| 2026-04-09 | Codex 互換性調査完了 — デュアルコーデック方式を採用 |
