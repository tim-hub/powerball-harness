# GitHub Harness Plugin Benchmark

最終更新: 2026-03-06

この文書は、GitHub で人気のある **Claude Code 向けハーネス / ワークフロープラグイン** を対象に、`claude-code-harness` を **導入後の標準運用がどう変わるか** という観点で比較した日付付きスナップショットです。

- これは **人気投票** ではなく **ハーネス比較** です
- GitHub stars は「比較対象の選定理由」としてだけ扱います
- まず「導入後に何が標準になるか」を並べ、その後で違いの意味を説明します
- 一般的な AI coding agent（Aider, OpenHands など）や curated list は、**単体ハーネスではない**ためこの比較表から外しています

## Compared Repositories

2026-03-06 時点で、GitHub 上で公開されており、かつ「Claude Code 向けの多段ワークフロー / プラグイン / ハーネス」を主張している repo のうち、比較に十分な公開情報があるものを対象にしました。

| Repo | GitHub stars | Included because |
|------|--------------|------------------|
| [obra/superpowers](https://github.com/obra/superpowers) | 71,993 | もっとも人気の高い workflow / skills 系プラグイン。比較対象として外せない |
| [gotalab/cc-sdd](https://github.com/gotalab/cc-sdd) | 2,770 | 要件駆動の開発フローを前面に出す人気の Claude Code 系ハーネス |
| [Chachamaru127/claude-code-harness](https://github.com/Chachamaru127/claude-code-harness) | 232 | 本 repo |

## ユーザーが見て分かる比較表

凡例:

- `✅` 導入直後から標準フローとして使える
- `△` 工夫すればできるが、主導線ではない
- `—` 主な訴求ではない

| ユーザーが気にすること | Claude Harness | Superpowers | cc-sdd |
|------------------------|----------------|-------------|--------|
| 計画が会話で消えずにリポジトリへ残る | ✅ | ✅ | ✅ |
| 実装が承認後に同じ流れで進みやすい | ✅ | ✅ | △ |
| レビューが完了前の標準工程に入る | ✅ | ✅ | △ |
| 危険な操作を実行時ガードで止める | ✅ | △ | — |
| 検証をあとから同じ手順でやり直せる | ✅ | △ | ✅ |
| 承認後は一気通貫で最後まで進められる | ✅ | △ | — |

## この違いが意味すること

### Claude Harness

- いちばん強いのは **標準フローの固定化** と **実行時ガード** と **再実行できる検証** です
- Plan → Work → Review が独立した導線として揃っていて、`/harness-work all` という一括実行の近道まであります
- 「毎回いい感じにやって」ではなく、「毎回同じ型で崩れず進んでほしい」人に向いています

### Superpowers

- いちばん強いのは **ワークフローの広さ** と **導入ストーリーの分かりやすさ** です
- 企画、実装、レビュー、デバッグまでの流れが見えやすく、自動トリガーも強いです
- ただし、危険な操作を実行時ルールで止める仕組みや、再実行できる証跡は Harness ほど標準フローとしては前面に出ていません

### cc-sdd

- いちばん強いのは **仕様駆動の規律** です
- `Requirements -> Design -> Tasks -> Implementation` の流れが明快で、dry-run や validate-gap / validate-design もあります
- ただし、公開面からは独立したレビュー工程や一括実行の導線が、Harness ほど標準フローとして強くは見えません

## README での見せ方

README や LP では、次の言い方が自然です。

> ワークフローの引き出しを広げたいなら Superpowers。
> 要件 → 設計 → タスクの規律を強めたいなら cc-sdd。
> 計画・実装・レビュー・検証を、崩れにくい標準フローに変えたいなら Claude Harness。

## 判定メモ

- `計画が会話で消えずにリポジトリへ残る`
  - Harness: `Plans.md` / `/harness-plan`
  - Superpowers: brainstorming / writing-plans workflow
  - cc-sdd: requirements / design / tasks workflow
- `実装が承認後に同じ流れで進みやすい`
  - Harness: `/harness-work --parallel`, Breezing, worker/reviewer flows が標準フローに乗る
  - Superpowers: parallel agent execution / subagent workflows が公開面で分かりやすい
  - cc-sdd: Claude agent variant では複数 subagent が確認できるが、すべての使い方で中心機能として打ち出されているわけではない
- `レビューが完了前の標準工程に入る`
  - Harness: `/harness-review` と `/harness-work all`
  - Superpowers: code review workflow is explicit
  - cc-sdd: validate コマンドは明示されているが、コードレビューを独立した工程として前面に出している度合いはやや弱い
- `危険な操作を実行時ガードで止める`
  - Harness: TypeScript guardrail engine + deny / warn rules
  - Superpowers: workflow discipline and hooks are visible, but compiled deny / warn runtime engine is not front-and-center
  - cc-sdd: 公開 README では、明示的な実行時 safety engine は確認しにくい
- `検証をあとから同じ手順でやり直せる`
  - Harness: validate scripts + consistency checks + evidence pack
  - Superpowers: verify-oriented workflows はあるが、artifact pack は前面に出ていない
  - cc-sdd: dry-run / validate-gap / validate-design がある
- `承認後は一気通貫で最後まで進められる`
  - Harness: `/harness-work all`
  - Superpowers: auto-triggered workflow はあるが、同じ意味での published single command は前面に出ていない
  - cc-sdd: spec-based command set はあるが、approval 後に full loop をまとめる単一の導線は前面に出ていない

## 注意点

- stars は毎日変わるため、この表は **日付付きスナップショット** です
- この比較は「市場人気」ではなく「ユーザーに見えるハーネス機能差」に寄せています
- `Superpowers > Claude Harness` となる軸もあります。特に ecosystem / adoption / workflow story の強さは目立ちます
- `cc-sdd > Claude Harness` となる軸もあります。特に要件駆動の規律の明快さは強みです
- README に載せるときは、勝ち負けの断言より **何を重視する人に向いているか** を書く方が自然です

## Evidence Used

### Local evidence

- [README.md](../README.md)
- [docs/claims-audit.md](claims-audit.md)
- [docs/distribution-scope.md](distribution-scope.md)
- [docs/evidence/work-all.md](evidence/work-all.md)

### Public GitHub sources

- [obra/superpowers](https://github.com/obra/superpowers)
- [gotalab/cc-sdd](https://github.com/gotalab/cc-sdd)
- [Chachamaru127/claude-code-harness](https://github.com/Chachamaru127/claude-code-harness)
