# Distribution Scope

最終更新: 2026-03-06

この文書は `claude-code-harness` の「repo には存在するが、常に同じ形で配布されるとは限らないもの」を明文化するための scope table です。
`Plans.md`、README、配布スクリプト、検証スクリプトで迷ったら、この表を正本として扱います。

## Scope Table

| Path | Status | Why it exists | Enforcement signal |
|------|--------|---------------|--------------------|
| `core/` | Distribution-included | TypeScript guardrail engine の本体 | `core npm test`, README architecture |
| `skills-v3/` | Distribution-included | 現行の 5 verb skills | README, mirror sync checks |
| `agents-v3/` | Distribution-included | 現行の worker / reviewer / scaffolder | README, validate-plugin-v3 |
| `hooks/` | Distribution-included | 実行時 guardrail と lifecycle hook | `hooks/hooks.json`, validate-plugin |
| `templates/` | Distribution-included | project init と rules 配布の正本 | `check-consistency.sh` |
| `commands/` | Compatibility-retained | 旧 slash command 資産。互換確認と mirror/build のため保持 | `tests/validate-plugin.sh`, `scripts/build-opencode.js` |
| `skills/` | Compatibility-retained | 旧 skill 群。移行済みだが既存導線の互換用に保持 | README architecture, codex mirror tests |
| `agents/` | Compatibility-retained | 旧 agent 群。移行済み導線の互換用に保持 | README architecture |
| `codex/`, `opencode/` | Distribution-included | 代替クライアント向け mirror / setup 導線 | `tests/test-codex-package.sh`, `opencode-compat.yml` |
| `mcp-server/` | Development-only and distribution-excluded | オプション機能。repo では開発・調査用に残すが配布 payload には含めない | `.gitignore`, CHANGELOG history |
| `harness-ui/`, `harness-ui-archive/` | Development-only and distribution-excluded | optional UI 実験・旧実装の保管 | `.gitignore`, CHANGELOG history |
| `docs/research/`, `docs/private/` | Private reference | 比較メモ、調査記録、公開前の下書き | repo reference only |

## Current Decisions

- `commands/` は削除済み扱いにしない。現在は **Compatibility-retained**。
- `mcp-server/` は削除済み扱いにしない。現在は **Development-only and distribution-excluded**。
- README や `Plans.md` で「削除」と書く場合は、実際に tree から消えたときだけ使う。
- 「配布外」「互換維持」「開発専用」はこの文書のラベルに合わせて使い分ける。

## Update Rule

次のいずれかが起きたら、この表も同じ PR / commit で更新すること。

1. README の architecture / install / compatibility 説明を変更したとき
2. `.gitignore` や build script の除外規則を変更したとき
3. `commands/` や `mcp-server/` など、存在理由が誤解されやすいディレクトリの扱いを変えたとき
