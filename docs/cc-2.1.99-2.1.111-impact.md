# Claude Code 2.1.99-2.1.111 影響整理

ひとことで:
Phase 44 の docs 判断として、`2.1.99` から `2.1.111` までの公開 changelog を棚卸しし、Harness への影響を `A` と `C` に分類した結果、`B` は `0` 件です。

たとえると:
新しい道具箱が届いたときに、「そのまま使える道具」と「こちらで持ち手や収納を作り直す必要がある道具」を仕分けしたメモです。

## 前提

- 一次情報は Anthropic 公開 `claude-code` リポジトリの `CHANGELOG.md` を使用
- この文書の分類は Phase 44 の計画と対応付けるためのもの
- `A`: Harness 側で明示的な追従が必要な項目
- `C`: Claude Code 本体の更新だけで恩恵を受ける項目
- `B`: 「書いただけ」は今回 `0`

## 要約

| 範囲 | 判定 | 補足 |
|------|------|------|
| `2.1.99`, `2.1.100`, `2.1.102`, `2.1.103`, `2.1.104`, `2.1.106` | `C` | 公開 changelog 上の個別項目なし。Phase 44 では追記対象なし |
| `2.1.101` | `C` | UX 改善と安定性修正が中心で、Harness 固有のコード追加は不要 |
| `2.1.105` | `A` | `PreCompact` hook と `monitors` manifest は Harness 側の統合が必要 |
| `2.1.107` | `C` | thinking 表示改善。Harness は自動継承 |
| `2.1.108` | `A` / `C` 混在 | 1 時間 prompt cache は明示追従対象、他は主に自動継承 |
| `2.1.109` | `C` | thinking indicator 改善のみ |
| `2.1.110` | `A` / `C` 混在 | permission 再評価まわりは明示追従対象、他は主に自動継承 |
| `2.1.111` | `A` / `C` 混在 | `xhigh`、`/ultrareview`、Auto Mode flag 廃止は正式追従対象 |

## バージョン別一覧

| Version | 主要変更 | Harness 影響 | 分類 | Phase 44 トレース |
|---------|----------|--------------|------|-------------------|
| `2.1.99` | 公開 changelog 上の個別項目なし | 範囲起点として確認のみ。Harness 独自対応なし | `C` | - |
| `2.1.100` | 公開 changelog 上の個別項目なし | 追加追従なし | `C` | - |
| `2.1.101` | `/team-onboarding`、OS CA trust、`/ultraplan` 初期環境自動化、resume 安定化など | 既存 workflow がそのまま恩恵を受ける。Phase 44 での新規コードは不要 | `C` | - |
| `2.1.102` | 公開 changelog 上の個別項目なし | 追加追従なし | `C` | - |
| `2.1.103` | 公開 changelog 上の個別項目なし | 追加追従なし | `C` | - |
| `2.1.104` | 公開 changelog 上の個別項目なし | 追加追従なし | `C` | - |
| `2.1.105` | `PreCompact` hook、plugin `monitors` manifest、`/proactive` alias など | `hooks.json` と plugin manifest に実装統合が必要 | `A` | `44.2.1`, `44.2.2` |
| `2.1.106` | 公開 changelog 上の個別項目なし | 追加追従なし | `C` | - |
| `2.1.107` | thinking indicator 改善 | 表示改善を自動継承 | `C` | - |
| `2.1.108` | `ENABLE_PROMPT_CACHING_1H`、recap、built-in slash command discovery など | 1 時間 cache は運用ポリシー整備が必要。他は概ね自動継承 | `A/C` | `44.6.1`, `44.7.1` |
| `2.1.109` | extended-thinking indicator 改善 | UI 上の恩恵のみ。追従コード不要 | `C` | - |
| `2.1.110` | permission deny 再評価 fix、`PreToolUse.additionalContext` fix、`/tui`、resume/scheduled task など | guardrail 再検証と docs 更新が必要。その他 UX 改善は自動継承 | `A/C` | `44.3.1`, `44.11.1` |
| `2.1.111` | `xhigh`、`/ultrareview`、Auto mode no longer requires flag、`/effort` slider など | `xhigh` と `/ultrareview` は正式追従対象。Auto Mode の前提文言も更新が必要 | `A/C` | `44.5.1`, `44.8.1`, `44.11.1` |

## 重点メモ

### `2.1.105`

- `PreCompact` hook は Harness の長時間実行保護に直結します
- `monitors` manifest は monitor 系スクリプトを「後付け」ではなく「起動時自動 arm」に変える起点です
- どちらも Harness の付加価値を作るので `A` です

### `2.1.108`

- `ENABLE_PROMPT_CACHING_1H` は、単に使えるだけではなく「どのフローで有効化するか」の方針が必要です
- そのため docs と policy を伴う `A` に寄せます
- 一方で recap や slash command discovery の多くは本体恩恵なので `C` です

### `2.1.110`

- `permissions.deny` の再評価修正は、Harness の guardrail 説明と期待動作に直接効きます
- 「CC が直したから終わり」ではなく、Harness の説明とテスト観点を更新する必要があるため `A/C` 混在です

### `2.1.111`

- `xhigh` は見送りではなく正式対象です
- `/ultrareview` も見送りではなく正式対象です
- `Auto mode no longer requires --enable-auto-mode` は、Auto Mode の docs 前提を古いままにしないための明示追従対象です

## B がゼロである理由

- 公開 changelog にない版は「無理に意味づけして行を増やさない」
- 本体改善だけで終わるものは `C` に寄せる
- Harness の判断・文言・設定・フックに影響するものだけを `A` に寄せる

この切り分けにより、「Feature Table に書いただけ」の `B` を作らずに済みます。

## 具体例

具体例:
`2.1.111` の `xhigh` は、単に新しい effort 名が増えた話ではありません。Harness 側では reviewer/advisor の thinking 強度ポリシーや docs の説明に影響するため、`A` として扱います。

## なぜこの整理にしたか

Phase 44 は「CC の changelog を要約する」ことが目的ではなく、「Harness がどこを自分で持つか」を明確にすることが目的だからです。
