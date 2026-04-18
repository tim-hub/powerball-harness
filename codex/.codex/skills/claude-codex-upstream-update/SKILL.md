---
name: claude-codex-upstream-update
description: "Claude Code changelog と Codex releases を調査し、Harness に実装価値のある差分だけを取り込むためのローカル専用スキル。Use when checking upstream updates, integrating Claude updates into Harness, comparing Claude vs Codex upgrade impact, or preparing the next upstream adaptation cycle. Do NOT use for public release copy or generic changelog summarization."
description-ja: "Claude Code changelog と Codex releases を調査し、Harness に実装価値のある差分だけを取り込むためのローカル専用スキル。公開配布しない。"
description-en: "Local-only skill for researching Claude Code and Codex upstream releases, selecting high-value adaptations, and implementing meaningful Harness upgrades. Not for distribution."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob", "WebSearch"]
user-invocable: false
---

# Claude / Codex Upstream Update

Claude Code と Codex の upstream 更新を、Harness の実装差分まで落とし込むためのローカル専用スキル。
目的は「更新を紹介すること」ではなく、「Harness を実際に強くすること」。

## 使う場面

- Claude Code の changelog を見て、Harness に取り込むべき更新を選びたいとき
- Codex releases を見て、Claude から使う場合 / Codex から使う場合の差を整理したいとき
- docs の追記だけで終わらせず、hooks / scripts / tests / Plans / CHANGELOG まで反映したいとき

## 使わない場面

- 公開向けの単純なリリースまとめ
- ただの changelog 要約
- 実装差分を伴わない宣伝文

## 基本ルール

1. 一次情報を先に確認する
   - Claude: `https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md`
   - Codex: `https://github.com/openai/codex/releases`
2. 既存の Harness 実装を確認する
   - `docs/CLAUDE-feature-table.md`
   - `skills/cc-update-review/SKILL.md`
   - `hooks/hooks.json`
   - `.claude-plugin/hooks.json`
   - `scripts/hook-handlers/`
   - `core/src/guardrails/`
   - `tests/test-claude-upstream-integration.sh`
   - `tests/validate-plugin.sh`
3. Feature Table に書くだけで終わらせない
   - `skills/cc-update-review/SKILL.md` の A/B/C 分類に従う
   - `B = 書いただけ` は不可。実装案か実装本体まで進める
4. 優先順位は Claude 側を先に決める
   - 今回すぐ実装するもの
   - Codex 比較軸として残すもの
   - CC / Codex の自動継承で済むもの

## 実行フロー

### 1. 公式更新を分解する

- Claude 側は「新機能」「運用改善」「修正」「自動継承」に分ける
- Codex 側は「今すぐ比較軸になるもの」と「将来タスク」を分ける
- ここで大事なのは、Harness が増幅できるかどうか

判断の目安:

- 実装対象
  - hook / script / skill / agent / validate に落とせる
  - 体験改善を「今まで / 今後」で説明できる
- 比較軸
  - 今回は実装しないが、Claude と Codex の差を埋める価値が高い
- 自動継承
  - upstream の修正だけで恩恵を受ける

### 2. 既存導線に載せる

更新候補は、次の面のどこに入るかを必ず決める。

- `hooks/` / `.claude-plugin/hooks.json`
- `scripts/hook-handlers/`
- `core/src/guardrails/`
- `skills/` / `agents/`
- `tests/test-claude-upstream-integration.sh`
- `tests/validate-plugin.sh`
- `docs/CLAUDE-feature-table.md`
- `CHANGELOG.md`
- `Plans.md`

「どこにも入らない」ものは、原則として実装対象にしない。

### 3. Claude 優先で実装する

優先するのは次のどちらか。

- Claude の新機能を Harness の既存導線に吸収すると、ノイズ削減・安全性向上・自動化強化が起こる
- Harness の既存機能が upstream 更新でさらに強くなる

例:

- hooks conditional `if` field を使って、permission hook を必要な Bash だけに絞る
- `MultiEdit` など既存 guardrail と hooks の不整合を埋める
- 新しい hook event を runtime tracking / recovery / Plans 再読リマインドへつなぐ

### 4. Codex は比較軸として残す

Codex 側は今回の主実装対象でなくても、次のような形で残す。

- `plugin-first workflow`
- `resume-aware effort continuity`
- readable agent addressing
- image-aware workflow

1 回の更新サイクルで全部やろうとしない。
Claude 側で価値が大きいものを先に完了させ、Codex 側は Plans に切り出す。

### 5. 記録を更新する

最低限、次を更新する。

- `Plans.md`
- `docs/CLAUDE-feature-table.md`
- `CHANGELOG.md`
- `tests/test-claude-upstream-integration.sh`
- 必要なら `tests/validate-plugin.sh`

書き方の基準:

- 何が upstream で増えたか
- Harness は何を取り込んだか
- ユーザー体験がどう変わるか

## 完了条件

- Claude / Codex の一次情報確認が済んでいる
- Claude 側で少なくとも 1 件は実装または検証強化まで終わっている
- Feature Table と CHANGELOG が「意味のある改善」として読める
- Plans に将来対応が残っている
- `validate-plugin` 系または対象テストを実行している

## 非配布ルール

- このスキルはローカル専用
- mirror しない
- 公開パッケージへの収録前提で設計しない
- 配布したくなった場合は、ローカル事情を抜いた別スキルとして切り出す

## 保存しておく実装メモ

### `PreToolUse updatedInput` を使った AskUserQuestion 自動補完・入力正規化

Claude Code 側で `PreToolUse updatedInput` が安定して使える場面では、`AskUserQuestion` の入力を「質問前に軽く整える」設計が有効。
ここでの正規化は、ユーザーの意図を書き換えることではなく、選択肢の補完や曖昧表現の整列を先に済ませて、質問のやり直し回数を減らすことを指す。

#### 狙い

- 同じ質問を言い換えて何度も聞き直す回数を減らす
- `harness-plan create` のような対話フローで、短い回答からでも安全に次の質問へ進める
- `solo / team`, `patch / minor / major`, `scripted / exploratory` のような既知の選択肢を早い段階でそろえる

#### 向いている surface

- `harness-plan create`
- `harness-release`
- 将来 `request_user_input` 相当を持つ対話型 setup / review 導線

#### 最小実装案

1. `PreToolUse` で `AskUserQuestion` 呼び出しを検出する
2. `updatedInput` で以下だけを整える
   - 既知の同義語を canonical value に寄せる
   - 空欄時は安全な default 候補を補助文として追加する
   - 選択肢にない値は勝手に変換せず、そのまま質問へ戻す
3. 変更した場合は trace へ「何を補完したか」を短く残す

#### 正規化してよいもの

- `solo`, `single`, `個人` → `solo`
- `team`, `issue`, `github issue` → `team`
- `browser exploratory`, `探索`, `触って確認` → `exploratory`
- `browser scripted`, `playwright`, `手順固定` → `scripted`

#### 正規化してはいけないもの

- ユーザーが自由入力した固有名詞
- スコープの広さを変える要約
- セキュリティや権限に関わる yes/no 判断

#### 次回着手時の変更候補

- `.claude-plugin/hooks.json`
- `hooks/hooks.json`
- `scripts/hook-handlers/` 配下の `PreToolUse` handler
- `skills/harness-plan/SKILL.md`
- `skills/harness-release/SKILL.md`
- `tests/test-claude-upstream-integration.sh`

#### Done の目安

- `updatedInput` で補完された値と元入力の両方が追跡できる
- 意図を書き換えず、既知の選択肢の整列だけに作用する
- 対話回数が減る一方で、誤変換時は元入力へ安全に戻せる
