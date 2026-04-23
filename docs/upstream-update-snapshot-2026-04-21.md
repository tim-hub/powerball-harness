# Claude Code / Codex upstream snapshot - 2026-04-21

この snapshot は、2026-04-21 時点の公式 upstream を確認し、Claude Code Harness に直接取り込むべき項目と、今回は自動継承 / Plans 化に留める項目を分解したもの。

一次情報:

- Claude Code changelog: <https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md>
- Claude Code docs changelog: <https://code.claude.com/docs/en/changelog>
- OpenAI Codex releases: <https://github.com/openai/codex/releases>

## Version-by-version breakdown

| Version | Upstream item | Category | Harness surface | Action |
|---------|---------------|----------|-----------------|--------|
| Claude Code 2.1.116 | Large-session `/resume` is faster and handles dead-fork entries more efficiently | C | session resume/fork UX | Harness の session-control 変更は不要。大容量 transcript の resume guidance と矛盾しないことを確認 |
| Claude Code 2.1.116 | MCP startup defers `resources/templates/list` until first `@` mention | C | MCP / `@` mention guidance | Harness 側の MCP setup を変更しない。将来 MCP health watch を作る場合は deferred startup 前提にする |
| Claude Code 2.1.116 | `/reload-plugins` and background auto-update install missing marketplace dependencies | P | plugin setup / marketplace docs | Plugin dependency policy の説明と setup smoke を後続で見直す。現時点では Harness に dependency resolver を重ねない |
| Claude Code 2.1.116 | Sandbox auto-allow no longer bypasses dangerous-path safety for `rm` / `rmdir` | C | guardrail / Bash safety | Phase 51 の R05 guardrail と同じ方向。Harness 側は既存 test を維持し、CC 本体修正を自動継承 |
| Claude Code 2.1.116 | Agent frontmatter `hooks:` fire for main-thread agents via `--agent` | P | agents / skills docs | Harness agents が `--agent` main-thread execution で hook 前提を持つかを後続 audit に残す |
| Claude Code 2.1.116 | Bash tool shows GitHub API rate-limit hints for `gh` commands | P | ci / release / review skills | `gh` retry/backoff guidance を CI / release skills に反映する候補。今回は実装せず Plans 化 |
| Codex 0.122.0 | `/side` conversations and queued slash / `!` shell prompts while work is running | P | long-running work guidance | Harness loop / breezing の「作業中の横質問」UX に活かせる可能性あり。Codex-native skill audit と合わせて扱う |
| Codex 0.122.0 | Plan Mode can start implementation in a fresh context with context usage shown | P | `/plan-with-agent` / `/work --codex` handoff | Plan -> Work の context carry policy と比較する。すぐに Harness の phase model は変えない |
| Codex 0.122.0 | Plugin workflows add tabbed browsing, toggles, marketplace removal, remote/cross-repo/local sources | P | plugin mirror / setup policy | Harness plugin mirror policy と marketplace source policy の後続整理に回す |
| Codex 0.122.0 | Filesystem permissions add deny-read glob, managed deny-read, platform sandbox enforcement, isolated `codex exec` | P | sandbox / guardrail | Claude 側 `sandbox.network.deniedDomains` とは別軸。Codex mirror の sandbox policy として後続比較 |
| Codex 0.122.0 | Tool discovery and image generation default-on, higher-detail image handling | P | Codex mirror skill metadata | allowed-tools / image skill / tool discovery guidance の drift audit に残す |
| Codex 0.122.0 | App-server stale prompt dismissal and resume/fork token usage replay | C/P | session resume / heartbeat | Codex 本体 UX は自動継承。Harness heartbeat / resume summary との重複は後続で確認 |
| Codex 0.123.0-alpha.2 | Pre-release with thin release body | P | future compare | release body から推測実装しない。stable 化後または release notes が厚くなった時に再確認 |

## UX judgement

今回すぐに Harness へ実装するべき upstream 機能は少ない。Phase 51 で `AskUserQuestion.updatedInput` と Claude 2.1.113 hardening を実装済みのため、2.1.116 / Codex 0.122.0 は「既存実装と矛盾しないか」「後続の setup / Codex-native skill audit に落とすか」を見る回にするのが安全。

直接実装しない理由:

- Claude 2.1.116 の多くは Claude Code 本体の TUI / resume / plugin updater 改善で、Harness が wrapper を重ねると挙動差分や二重責務を作りやすい。
- Codex 0.122.0 は plugin workflow / filesystem permission / Plan Mode など大きい設計変更が多く、既存 Phase 51.2 の Codex-native skill audit と同時に扱う方が依存関係を壊しにくい。
- Codex 0.123.0-alpha.2 は pre-release かつ release body が薄いため、compare から推測実装しない。

## Follow-up candidates

- Codex plugin marketplace source policy と Harness mirror policy の統合案を Phase 51.2.3 と合わせて整理する。
- CI / release / review skills に `gh` rate-limit hint を受けた retry/backoff 方針を追加するか判断する。
- `--agent` main-thread hooks の挙動を Harness agents の frontmatter policy に反映するか確認する。
- Codex deny-read glob / isolated `codex exec` と Harness sandbox / guardrail policy の差分表を作る。
