# 長時間タスク実行ガイド

この文書は、Claude Code で **1 回では終わらない作業** を安全に回すための実務ガイドです。
ここでいう「長時間タスク」は、`/loop` と `ScheduleWakeup` を使って、少しずつ進める仕事のことです。
この文書は Phase 41.4.1 の成果物です。

対象は **Phase 41 の同一セッション内の運用** です。別ホストをまたいだ自動再入は、この段階では扱いません。

参考: [skills/harness-loop/SKILL.md](../skills/harness-loop/SKILL.md) / [skills/harness-loop/references/flow.md](../skills/harness-loop/references/flow.md) / [docs/CLAUDE-feature-table.md](CLAUDE-feature-table.md)

---

## 1. まず全体像をつかむ

長時間タスクは、次の 4 つをくり返して進めます。

1. 今やる 1 単位の作業を決める
2. 小さく実装または確認する
3. 結果を checkpoint として残す
4. 次の wake-up を予約する

ここで大事なのは、**毎回「新しい視点」で入り直す** ことです。
前回の会話をそのまま引きずるのではなく、resume pack で必要な情報だけを再注入して再開します。

### B1-B12 の 12 軸で見る対応表

| 軸 | 何を決めるか | Harness での対応 |
|---|---|---|
| B1 | 何を達成したいか | Plans.md の対象タスクと DoD を読む |
| B2 | 1 回でどこまでやるか | 1 サイクル = 1 タスク単位で進める |
| B3 | どこから始めるか | `/loop` を入口にする |
| B4 | どう再開するか | `ScheduleWakeup` で次回 wake-up を予約する |
| B5 | 何を引き継ぐか | `harness-mem resume-pack` で必要情報だけ戻す |
| B6 | どれくらい待つか | `pacing` で間隔を選ぶ |
| B7 | いつ止めるか | `--max-cycles` で上限を設ける |
| B8 | どう衝突を避けるか | lock と冪等性ガードで多重起動を防ぐ |
| B9 | どう進捗を残すか | `harness_mem_record_checkpoint` で checkpoint を記録する |
| B10 | うまく進んでいるか | plateau 検知で停滞を見つける |
| B11 | どこまでを対象にするか | Phase 41 は同一セッション内に限定する |
| B12 | 何に注意するか | `bypassPermissions` と Plans.md flock の限界を理解する |

---

## 2. `/loop` + `ScheduleWakeup` の使い方

`/loop` は、Claude Code に「作業を続ける前提」を伝える入口です。
`ScheduleWakeup` は、次の再開時刻を予約する仕組みです。

### 使い方の基本

```text
/loop all
/loop 41.1-41.3 --pacing ci
/loop all --pacing night
```

### 1 回の流れ

1. `Plans.md` から次の対象タスクを 1 件選ぶ
2. そのタスクに必要な最小作業だけを実行する
3. checkpoint を残す
4. 次の wake-up を `ScheduleWakeup` で予約する

### 予約のイメージ

```text
ScheduleWakeup(
  delaySeconds=270,
  prompt="/harness-loop all --cycles-done 1 --pacing worker",
  reason="1 サイクル完了。次のタスクに進むため"
)
```

`delaySeconds` は「何秒後に戻ってくるか」です。
短すぎると慌ただしく、長すぎると前回の流れを忘れやすくなります。
実際には 60 〜 3600 秒の範囲に収めます。

---

## 3. pacing プリセットの選び方

`pacing` は、次の wake-up をどれくらい空けるかの設定です。

| pacing | delaySeconds | 向いている場面 | ひとこと |
|---|---:|---|---|
| `worker` | 270 | 直前の作業からすぐ続けたい | 標準設定 |
| `ci` | 270 | CI の結果待ちがある | 待ち時間を短く保つ |
| `plateau` | 1200 | 進みが止まりやすい | 少し長めに冷ます |
| `night` | 3600 | 夜間にまとめて回したい | いちばん長い待機 |

### cache 境界の考え方

Claude Code には、直前の流れを短時間だけ覚えておける「短期キャッシュ」があります。
`worker` と `ci` の 270 秒は、この短期キャッシュにまだ乗りやすい長さです。

一方で `plateau` や `night` は、短期キャッシュが切れやすいので、**resume pack を必ず前提にする** のが安全です。
つまり、待ち時間が長いほど「自力で思い出す」のではなく「必要情報を再注入する」設計に寄せます。

### 1時間キャッシュを使う時

Claude Code `2.1.108` 以降では、`ENABLE_PROMPT_CACHING_1H=1` を付けると、
通常の 5 分キャッシュより長い **1 時間キャッシュ** を opt-in できます。

これは「毎回ほぼ同じ前提を読み直すが、次の入力が 5 分を超えやすい」時に向いています。
このドキュメントで扱う長時間タスクでは、特に次の場面と相性が良いです。

1. `/harness-loop` で 1 サイクルごとに待機が入る
2. `/resume` や `/continue` をまたいで同じ前提を使い回す
3. レビューや advisor consult をはさみ、5 分を超えて戻ることがある

逆に、数十秒から数分の短い往復が連続するだけなら、既定の 5 分キャッシュのままで十分です。

### 1h vs 5m cache の選択基準

| 判定軸 | 1h cache を選ぶ | 5m cache（既定）で足りる |
|--------|----------------|------------------------|
| セッション長の見込み | **30 分を超える** | 30 分以内 |
| wake-up 間隔 | `plateau`（1200s）や `night`（3600s） | `worker`/`ci`（270s） |
| 前提情報の再利用 | 毎サイクルほぼ同じ SKILL.md・Plans.md を読む | 前提が毎回変わる短い往復 |
| 対象スキル | `/breezing` / `/harness-loop` の多タスク実行 | 単発の `/work` や対話 |

**判定ルール**: セッション長が **30 分を超える見込み** なら 1h cache を選ぶ。それ以外は既定の 5 分 cache で十分。

opt-in 手順:

```bash
bash scripts/enable-1h-cache.sh
```

このコマンドは `env.local` に `ENABLE_PROMPT_CACHING_1H=1` を追記します（冪等）。
グローバル設定は変えません。すでに設定済みの場合は何もしません。

### 推奨導入方針

このリポジトリでは、**全セッション常時オンにはしません**。
理由は、1 時間キャッシュは便利ですが、追加コスト前提であり、短い対話には過剰になりやすいからです。

代わりに、長時間タスク専用の薄い起動ラッパーを使います。

```bash
bash scripts/claude-longrun.sh
```

そのまま引数も渡せます。

```bash
bash scripts/claude-longrun.sh --resume
bash scripts/claude-longrun.sh --model claude-opus-4-6
```

このスクリプトは、内部で `ENABLE_PROMPT_CACHING_1H=1` を付けて `claude` を起動するだけです。
グローバル設定は変えないので、通常作業への影響を広げません。

---

## 4. wake-up 回数上限・lock・冪等性ガード

長時間タスクは、気づかないうちに同じ処理を二重に走らせることがあります。
これを防ぐために、3 層で守ります。

### 4-1. 回数上限

`--max-cycles` で、何回まで続けるかを決めます。
上限に達したら、そこでいったん止めます。

### 4-2. lock

同じタスクが同時に 2 回動かないように、lock を取ります。
このリポジトリでは `.claude/state/locks/loop-session.lock.d` を使います。

lock は「ここは今すでに動いている」という目印です。
もし既に lock があれば、新しい実行は止めます。
これで、並行実行による競合を防ぎます。

### 4-3. 冪等性ガード

冪等性は、同じ操作を 2 回やっても壊れない性質です。
`tests/validate-plugin.sh --quick` のような軽い確認を先に入れることで、壊れた状態で無理に進まないようにします。

また、lock は終了時に必ず片付けます。
正常終了でも異常終了でも、残骸が次回の邪魔をしないようにするためです。

---

## 5. plateau 検知と golden fixture

plateau は、作業が進んでいるように見えて、実は同じ所をぐるぐる回っている状態です。
たとえば、同じ修正を何度も繰り返す、判断材料が増えないのに再実行だけ増える、というときに起こります。

### 閾値の考え方

実際の判定は `scripts/detect-review-plateau.sh` の結果で見ます。
ここでは「何回失敗したら止めるか」よりも、**新しい情報が増えているか** を重視します。

### 何を fixture にするか

回帰を防ぐための golden fixture は、`tests/fixtures/` 配下に置くのが分かりやすいです。
たとえば `tests/fixtures/long-running-harness/` のように、長時間タスク専用のまとまりにすると見つけやすくなります。
特に plateau 関連は、次のようなケースを固定化すると役に立ちます。

1. 失敗理由が毎回同じケース
2. 条件を変えても判定が変わらないケース
3. 一見進んでいるようで実際は停滞しているケース

fixture は「この判定が今後も同じであるべき」という見本です。
これがあると、後でロジックを触ったときに、停滞検知が壊れていないか確認しやすくなります。

---

## 6. Phase 41 のスコープ

この Phase 41 で扱うのは、**同じ Claude Code セッションの中で完結する長時間タスク** です。

やることは次の 2 点に絞ります。

1. いまのセッション内で安全に再入できること
2. wake-up をまたいでも、同じ作業を続けられること

やらないことは、別ホストをまたいだ自動再入です。
それは将来の Phase 42 以降で考える範囲です。

---

## 7. 既知の制約

### `bypassPermissions` との関係

`/loop` は、権限を増やす仕組みではありません。
既存の権限ガードがある前提で動きます。
つまり、`bypassPermissions` を使っていても、危険な操作が無制限になるわけではありません。

長時間タスクでは、むしろ「勝手に強いことをしない」ほうが大切です。
必要な操作だけを、必要なタイミングで、必要な回数だけ行います。

### Plans.md flock の限界

`Plans.md` は複数の実行主体が触ることがあります。
そこで flock で順番待ちをする設計になっていますが、これは **同じファイルを同時に書き壊さないための仕組み** であって、万能ではありません。

特に、別セッションや別プロセスが同時に読んでいると、見えている状態が少し遅れることがあります。
そのため、`Plans.md` を読むときは「今見えている内容が最新とは限らない」前提を持ち、checkpoint や contract と合わせて判断します。

---

## 8. すぐ見るリンク

- 実行フローの詳細: [skills/harness-loop/references/flow.md](../skills/harness-loop/references/flow.md)
- コマンド入口: [skills/harness-loop/SKILL.md](../skills/harness-loop/SKILL.md)
- Claude Code の機能一覧: [docs/CLAUDE-feature-table.md](CLAUDE-feature-table.md)
