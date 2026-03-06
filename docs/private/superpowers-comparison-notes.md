# Superpowers Comparison Notes

最終更新: 2026-03-06

このメモは公開 PR 用ではなく、README / LP / benchmark 更新時の内部根拠用です。

## Snapshot

- Comparison target: `obra/superpowers`
- Review date: 2026-03-06
- Local static snapshot path: `/tmp/superpowers-review.hMYeEg/repo`

## Reusable Findings

### Harness strengths

- runtime enforcement が厚い。hooks と TypeScript core が明確
- consistency check, validate-plugin, core test で「主張に対する検査導線」を作りやすい
- Codex / OpenCode / Claude の multi-client 配布導線が repo に織り込まれている

### Superpowers strengths

- public narrative が鋭い。README の約束が分かりやすい
- market adoption / social proof が非常に強い
- operator 目線では「何を学べばいいか」が直感的

### Harness weaknesses that matter

- README / Plans / docs の自己矛盾があると、実力差より大きく信用を失う
- 配布対象と repo 内残置物の境界が曖昧だと、保守コストの高さだけが先に見える
- `/harness-work all` のような強い claim は、再現証拠がないと narrative で負ける

## Recommended Direction

1. badge / docs / Plans drift を最優先で潰す
2. `/harness-work all` evidence pack を公開導線に乗せる
3. README を `5 verb skills + TypeScript guardrail engine` に再集中する
4. 競合批判ではなく、runtime enforcement と verification で差別化する
