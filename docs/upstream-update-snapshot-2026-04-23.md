# Claude Code / Codex upstream snapshot - 2026-04-23

この snapshot は、2026-04-23 時点の公式 upstream を確認し、Claude Code Harness に直接取り込むべき項目と、自動継承 / 将来タスクに留める項目を分解したもの。

確認日:

- 2026-04-23 (Asia/Tokyo)

一次情報:

- Claude Code docs changelog: <https://code.claude.com/docs/en/changelog>
- Claude Code GitHub changelog: <https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md>
- OpenAI Codex releases: <https://github.com/openai/codex/releases>
- OpenAI Codex `rust-v0.123.0` release tag: <https://github.com/openai/codex/releases/tag/rust-v0.123.0>

確認対象:

- Claude Code `2.1.117`
- Claude Code `2.1.118`
- Codex `0.123.0`

分類:

- `A: 実装`: Harness 側で hooks / release / setup / skills / tests / docs のいずれかに落とす。53.1.1 では実装タスク ID まで固定し、実装本体は後続 task で行う。
- `C: 自動継承`: Claude Code / Codex 本体の改善をそのまま受ける。Harness wrapper を重ねると二重責務になるもの。
- `P: 将来タスク`: 今回すぐに実装しないが、Plans に後続候補として残す。推測実装はしない。

## Version-by-version breakdown

| Version | Upstream item | こうよくなる | Category | Harness surface | Harness action |
|---------|---------------|--------------|----------|-----------------|----------------|
| Claude Code 2.1.118 | Vim visual mode / visual-line mode | キーボードだけで選択と編集がしやすくなる | C | TUI usage | Harness 側の wrapper は不要。Claude Code 本体の editor UX として自動継承 |
| Claude Code 2.1.118 | `/cost` と `/stats` が `/usage` に統合され、旧コマンドは relevant tab を開く typing shortcut になる | 利用量確認の入口が 1 つにまとまり、迷いにくくなる | C | docs / session guidance | 53.1.6 で古い `/cost` / `/stats` 中心の説明があれば `/usage` 中心へ寄せる。挙動は本体を自動継承 |
| Claude Code 2.1.118 | `/theme` で named custom themes を作成 / 切替でき、JSON 手編集と plugin `themes/` directory に対応 | plugin が見た目の初期値を配れる余地ができる | P | plugin setup / design policy | 53.1.5 で setup docs に用途を記録する。Harness plugin に theme を同梱するかは今回は決めず、推測実装しない |
| Claude Code 2.1.118 | Hooks can invoke MCP tools directly via `type: "mcp_tool"` | shell script を挟まない読み取り診断 hook を作れる可能性がある | A | hooks / MCP diagnostics / tests | 53.1.2 で読み取り専用 MCP health/resource 診断に限定して検証した。今回の配布 hooks manifest は no-op とし、書き込み系 MCP tool を hook から呼ばないことを test で固定 |
| Claude Code 2.1.118 | `DISABLE_UPDATES` blocks all update paths, including manual `claude update` | 企業管理環境で手動更新まで止められる | A | setup docs / plugin policy | 53.1.5 で `DISABLE_AUTOUPDATER` との差を説明する。Harness 独自 updater block は追加しない |
| Claude Code 2.1.118 | WSL can inherit Windows-side managed settings via `wslInheritsWindowsSettings` | Windows / WSL 混在環境の管理設定が揃いやすくなる | P | enterprise setup docs | 53.1.5 の managed settings 整理に含める。Harness default には過剰適用しない |
| Claude Code 2.1.118 | Auto Mode `autoMode.allow` / `soft_deny` / `environment` can include `"$defaults"` to extend built-ins | 既定安全ルールを消さずに独自ルールを足せる | A | permissions / sandbox docs / settings template | 53.1.4 で「置換」ではなく「`"$defaults"` へ追加」として guidance と test を固定する |
| Claude Code 2.1.118 | Auto mode opt-in prompt adds "Don't ask again" | 繰り返し確認のノイズが減る | C | Auto Mode UX | Claude Code 本体の対話 UX として自動継承。Harness 側で prompt suppression を重ねない |
| Claude Code 2.1.118 | `claude plugin tag` creates release git tags with plugin version validation | plugin release で version 不一致のまま tag を切りにくくなる | A | harness-release / release tests | 53.1.3 で release flow に preflight / dry-run guidance と validation 位置を追加する |
| Claude Code 2.1.118 | `--continue` / `--resume` find sessions that added the current directory via `/add-dir` | multi-directory session の再開漏れが減る | C | session-control docs | 53.1.6 で古い resume 前提があれば更新。実装は本体を自動継承 |
| Claude Code 2.1.118 | `/color` syncs session accent color to claude.ai/code when Remote Control is connected | remote UI の見た目が揃う | C | Remote Control UX | Harness surface なし。自動継承 |
| Claude Code 2.1.118 | `/model` picker honors `ANTHROPIC_DEFAULT_*_MODEL_NAME` / `_DESCRIPTION` overrides with custom gateways | gateway 環境で model 表示が実態に近づく | C | model guidance | Harness が model picker を包まないため自動継承。必要なら provider docs の後続で触れる |
| Claude Code 2.1.118 | Plugin auto-update skips caused by another plugin's version constraint appear in `/doctor` and `/plugin` Errors | plugin dependency の失敗理由が見つけやすくなる | A | setup / plugin policy docs | 53.1.5 で dependency auto-resolve / missing dependency hints と合わせて説明する |
| Claude Code 2.1.118 | MCP OAuth / custom header authentication fixes | MCP 接続の再認証ループや一時 401 後の詰まりが減る | C | MCP runtime | 本体修正を自動継承。Harness の MCP health watch は後続で、OAuth workaround は追加しない |
| Claude Code 2.1.118 | Credential save crash on Linux / Windows no longer corrupts `~/.claude/.credentials.json` | 認証情報破損のリスクが下がる | C | auth runtime | 本体修正を自動継承。Harness は credentials を直接書き換えない |
| Claude Code 2.1.118 | `/login` clears `CLAUDE_CODE_OAUTH_TOKEN` session token so disk credentials can take effect | env token 起動後の login が効きやすくなる | C | auth runtime | 本体修正を自動継承 |
| Claude Code 2.1.118 | New message scroll pill and `/plugin` badges readability fixes | TUI 表示が読みやすくなる | C | TUI | 自動継承 |
| Claude Code 2.1.118 | Plan acceptance dialog no longer offers auto mode when running with `--dangerously-skip-permissions` | permission mode 表示の混乱が減る | C | permission UX | 自動継承。Harness の permission guidance は 53.1.4 で `"$defaults"` に絞って更新 |
| Claude Code 2.1.118 | Agent-type hooks no longer fail with "Messages are required for agent hooks" on non-Stop events | agent hook の適用範囲が扱いやすくなる | C | agent hooks | 本体修正を自動継承。53.1.2 の `mcp_tool` hook とは別扱い |
| Claude Code 2.1.118 | `prompt` hooks no longer re-fire on tool calls made by an agent-hook verifier subagent | verifier subagent 起因の hook 再入ノイズが減る | C | hook runtime | 自動継承。Harness 側の再入防止 hook は追加しない |
| Claude Code 2.1.118 | `/fork` stores a parent pointer instead of writing the full parent conversation per fork | fork が軽くなり、disk 使用量が減る | C | session/fork | 自動継承。Harness session-state は pointer hydration を前提にできる |
| Claude Code 2.1.118 | Alt+K / Alt+X / Alt+^ / Alt+_ keyboard freeze fixes | キーボード入力が固まりにくくなる | C | TUI input | 自動継承 |
| Claude Code 2.1.118 | Remote session connect no longer overwrites local `model` setting | remote session 利用時の local 設定破壊が減る | C | remote session | 自動継承 |
| Claude Code 2.1.118 | Typeahead no longer errors when pasted file paths start with `/` | 絶対パス貼り付けが自然に使える | C | prompt input | 自動継承 |
| Claude Code 2.1.118 | `plugin install` on already-installed plugin re-resolves wrong-version dependencies | 依存関係の修復がしやすくなる | A | plugin setup docs | 53.1.5 で Harness 独自 resolver を重ねず本体に任せる方針として記録 |
| Claude Code 2.1.118 | File watcher invalid path / fd exhaustion errors are handled | 長時間起動時の watcher エラーで落ちにくくなる | C | long-running sessions | 自動継承 |
| Claude Code 2.1.118 | Remote Control sessions are not archived on transient CCR initialization blips | 一時的な remote 初期化失敗で session が消えにくくなる | C | remote session | 自動継承 |
| Claude Code 2.1.118 | Subagents resumed via `SendMessage` restore the explicit `cwd` they were spawned with | subagent resume 後も作業ディレクトリがずれにくい | C | subagent orchestration | Codex native `send_input` とは別の Claude Code 本体修正として自動継承 |
| Claude Code 2.1.117 | Forked subagents can be enabled on external builds with `CLAUDE_CODE_FORK_SUBAGENT=1` | external build でも forked subagent を試せる | P | agents / skills docs | 53.1.6 で将来候補として整理。Harness default に環境変数を強制しない |
| Claude Code 2.1.117 | Agent frontmatter `mcpServers` load for main-thread agent sessions via `--agent` | main-thread agent でも MCP 前提の agent 設定が効きやすくなる | P | agents audit / MCP setup | 53.1.6 で agents audit の後続候補として記録する |
| Claude Code 2.1.117 | `/model` selections persist across restarts and startup header shows project / managed-settings pins | model 設定の出どころが分かりやすくなる | C | model guidance | 自動継承。Harness は model pin を上書きしない |
| Claude Code 2.1.117 | `/resume` offers to summarize stale large sessions before re-reading | 大きく古い session の再開が軽くなる | C | session-memory / resume docs | 53.1.6 で wrapper を足さない理由を記録。Claude Code 本体の summary を優先 |
| Claude Code 2.1.117 | Faster MCP startup when local and claude.ai MCP servers are both configured | startup 待ち時間が短くなる | C | MCP startup | 自動継承。MCP health watch は 53.1.2 の読み取り診断候補に限定 |
| Claude Code 2.1.117 | Already-installed plugin install now installs missing dependencies | dependency 抜けの自己修復がしやすくなる | A | plugin setup docs | 53.1.5 で auto-resolve に Harness resolver を重ねない方針として記録 |
| Claude Code 2.1.117 | Plugin dependency errors include install hints and `claude plugin marketplace add` auto-resolves missing dependencies | marketplace 由来の依存解決が分かりやすくなる | A | plugin setup / marketplace policy | 53.1.5 で企業利用・安全な marketplace 運用として整理 |
| Claude Code 2.1.117 | Managed settings `blockedMarketplaces` / `strictKnownMarketplaces` are enforced across plugin install/update/refresh/autoupdate | 管理対象 marketplace policy が抜けにくくなる | A | managed settings docs | 53.1.5 で通常ユーザー default へ過剰適用しない形で説明 |
| Claude Code 2.1.117 | Advisor Tool experimental label / link / startup notification, plus stuck-result fixes | experimental Advisor の位置づけが分かりやすくなり、詰まりが減る | C | advisor strategy | Harness Advisor Strategy は現状維持。最終品質判定は Reviewer に残す |
| Claude Code 2.1.117 | `cleanupPeriodDays` retention sweep covers tasks, shell snapshots, backups | 古い補助データが溜まりにくくなる | C | maintenance / session storage | 本体 cleanup を自動継承。Harness maintenance で重複削除しない |
| Claude Code 2.1.117 | OpenTelemetry command / usage / effort attributes | 観測データが詳しくなる | C | telemetry | Harness が OTEL schema を直接扱っていないため自動継承 |
| Claude Code 2.1.117 | Native macOS/Linux builds replace Glob/Grep tools with embedded `bfs` and `ugrep` through Bash | 検索が速くなり、tool round-trip が減る | C | search guidance | 53.1.6 で wrapper を追加しない項目として整理 |
| Claude Code 2.1.117 | Windows caches `where.exe` executable lookups per process | Windows subprocess 起動が速くなる | C | Windows runtime | 自動継承 |
| Claude Code 2.1.117 | Default effort for Pro/Max subscribers on Opus 4.6 / Sonnet 4.6 is `high` | 複雑な作業の初期品質が上がりやすい | C | effort guidance | 53.1.6 で古い medium 前提があれば更新。Harness が無理に上書きしない |
| Claude Code 2.1.117 | OAuth / WebFetch / proxy / keyboard / SDK reload / Bedrock / MCP elicitation / subagent model / idle memory / VS Code plugin panel / Opus context fixes | 認証、fetch、proxy、入力、MCP、subagent、memory、context 表示の不具合が減る | C | runtime stability | 本体修正を自動継承。Harness workaround は追加しない |
| Codex 0.123.0 | Built-in `amazon-bedrock` model provider with configurable AWS profile support | Codex 側でも Bedrock provider を標準導線で扱いやすくなる | A | Codex setup docs / provider policy | 53.2.1 で provider guidance を更新し、Claude 側 Bedrock guidance と混ぜない |
| Codex 0.123.0 | `/mcp verbose` shows diagnostics, resources, and resource templates while plain `/mcp` stays fast | 困った時だけ詳しい MCP 診断を見られる | A | troubleshoot / setup skill | 53.2.2 で通常 `/mcp` と verbose 診断の使い分けを記録 |
| Codex 0.123.0 | Plugin MCP loading accepts both `mcpServers` and top-level server maps in `.mcp.json` | 既存 plugin MCP 設定の形をより広く読める | A | Codex setup / plugin MCP docs | 53.2.2 で両形式を説明し、Claude Code 側用語と混ぜない |
| Codex 0.123.0 | Realtime handoffs let background agents receive transcript deltas and explicitly stay silent | 長時間作業中の background agent が必要な時だけ返答しやすい | A | harness-loop / breezing guidance | 53.2.3 で silence policy を整理する。advisor/reviewer drift 検知とは矛盾させない |
| Codex 0.123.0 | Host-specific `remote_sandbox_config` requirements for remote environments | remote 環境ごとの sandbox 要件を分けられる | A | sandbox / execution policy | 53.2.4 で比較表化し、既存 `--approval-policy` / `--sandbox` guidance との重複を確認 |
| Codex 0.123.0 | Bundled model metadata refreshed, including current `gpt-5.4` default | 現在の既定 model 情報に追従しやすくなる | A | Codex setup docs | 53.2.1 で Harness 側が固定しすぎない方針として記録 |
| Codex 0.123.0 | `/copy` after rollback copies the latest visible assistant response | rollback 後のコピー内容が直感どおりになる | C | TUI UX | 53.2.5 で自動継承 bug fix として記録。workaround は追加しない |
| Codex 0.123.0 | Follow-up text submitted during manual shell commands is queued | manual shell 中に次入力しても `Working` stuck が起きにくい | C | long-running UX | 53.2.5 で長時間作業 UX に効く自動継承として記録 |
| Codex 0.123.0 | Unicode / dead-key input fixed in VS Code WSL terminals | WSL terminal で日本語・記号入力が壊れにくくなる | C | terminal input | 53.2.5 で自動継承として記録 |
| Codex 0.123.0 | Stale proxy env vars are not restored from shell snapshots | 古い proxy 設定で通信が壊れにくくなる | C | session shell snapshots | 53.2.5 で自動継承として記録 |
| Codex 0.123.0 | `codex exec` inherits root-level shared flags such as sandbox and model options | wrapper 側で同じ flag を重ねる必要が減る可能性がある | A | codex exec wrapper / sandbox docs | 53.2.4 で重複フラグ削減可否を確認する |
| Codex 0.123.0 | Review prompts no longer leak into TUI transcripts | review の内部プロンプトが見えにくくなる | C | review privacy | Codex 本体修正を自動継承 |
| Codex 0.123.0 | Code Review skill instructions tightened | Codex-driven review の指示が堅くなる | C | review skill | Harness reviewer は別 surface。必要なら Phase 53.3.1 で重複を整理 |
| Codex 0.123.0 | App-server protocol docs updated for threadless MCP resource reads and namespaced dynamic tools | MCP resource read / dynamic tool の説明が増える | P | future MCP / app-server docs | 53.2.2 の scope 外で必要になったら扱う。release body 以上の推測実装はしない |
| Codex 0.123.0 | Dependency alerts fixed, Rust dev debug-info reduced, Python app-server SDK types refreshed | 配布物と開発体験が安定する | C | dependency / build runtime | Harness 側の直接変更なし |

## 53.1.2 MCP tool hook decision

対象 hook の用途:

- 将来の読み取り専用の MCP health / resource list 診断に限定する。
- たとえば、MCP server の疎通状態、公開 resources の一覧、resource template の有無などを、shell script を増やさず hook から直接確認する用途。
- 外部状態を書き換える audit log、checkpoint 記録、issue 作成、database 更新、ファイル書き込みは対象外。

今回の判断:

- `hooks/hooks.json` / `.claude-plugin/hooks.json` は今回は no-op として変更しない。
- 理由は、2026-04-23 時点で公式 changelog は `type: "mcp_tool"` を告知している一方、Hooks reference の hook handler field 表はまだ `command` / `http` / `prompt` / `agent` 中心で、`mcp_tool` の必須 field 名と入力展開規約が安定仕様として読み取れないため。
- もう 1 つの理由は、配布 plugin の hooks は有効化された全環境で読まれるため、常に存在する読み取り専用 MCP diagnostic tool をまだ前提にできないため。環境依存の MCP tool を manifest に直接入れると、未設定環境で hook error を増やす可能性がある。

安全条件:

- 書き込み系 MCP tool は hook から呼ばない。
- 将来 `type: "mcp_tool"` を manifest に入れる場合でも、tool 名は `health` / `list` / `read` / `get` / `status` / `diagnostic` / `resource` など読み取り診断と分かるものに限定する。
- `write` / `create` / `update` / `delete` / `remove` / `record` / `mutate` / `set` / `insert` / `upsert` / `patch` を含む tool 名は hook から呼ばない。
- この方針は `tests/test-claude-upstream-integration.sh` で固定する。現時点は no-op を検出し、将来 `mcp_tool` hook が追加された場合は読み取り専用名であることを jq check する。

## 53.1.3 plugin tag release flow decision

対象:

- Claude Code `2.1.118` の `claude plugin tag`
- `harness-release` の release flow

今回の判断:

- `skills/harness-release/SKILL.md` に Claude plugin project 用の tag preflight を追加する。
- `.claude-plugin/plugin.json` がある project では、`VERSION` と `.claude-plugin/plugin.json` の version が一致しない限り tag に進まない。
- Pre-Gate と `--dry-run` では `claude plugin tag .claude-plugin --dry-run` を実行し、作られる plugin tag 名と内部の tag / push 相当コマンドを見える化する。
- Post-Gate では release commit 後に再度 version sync を確認し、`claude plugin tag .claude-plugin --push --remote origin` で `{plugin-name}--v{version}` tag を作る。
- 既存の GitHub Release automation が `vX.Y.Z` tag を前提にしている project では、plugin tag とは別に semver tag を作る。plugin 配布 tag は `claude plugin tag` に任せ、手動 `git tag` だけに依存しない。

安全条件:

- `claude plugin validate .claude-plugin/plugin.json` が失敗した場合は tag を作らない。
- `VERSION` と `.claude-plugin/plugin.json` が不一致の場合は tag を作らない。
- `--dry-run` は tag を作らず、release plan に実行コマンドを表示する目的で使う。
- この guidance は `tests/test-claude-upstream-integration.sh` で grep 固定する。

## 53.1.4 Auto Mode "$defaults" permission and sandbox policy

対象:

- Claude Code `2.1.118` の Auto Mode `autoMode.allow` / `autoMode.soft_deny` / `autoMode.environment`
- Harness の `.claude-plugin/settings.json`
- Project-level template: `templates/claude/settings.security.json.template`

今回の判断:

- Auto Mode built-in defaults stay in place through "$defaults"。
- Harness は、Claude Code 組み込みの Auto Mode default を置き換えない。
- Project / enterprise 側で `autoMode.allow` / `autoMode.soft_deny` / `autoMode.environment` を追加する場合だけ、各配列に `"$defaults"` を入れ、その後ろに project-specific / organization-specific entry を足す。
- 配布 plugin の `.claude-plugin/settings.json` には、この task では `autoMode` object を追加しない。理由は、組み込み default の中身と更新責務は Claude Code 本体が持つべきで、Harness が空の置換設定を配ると upstream default とのズレを作りやすいため。
- Project-level template には、`autoMode` を追加する時の注意書きだけを置く。実際の追加 entry はプロジェクトごとに異なるため、Harness が推測して固定しない。

追加時の形:

```json
{
  "autoMode": {
    "allow": ["$defaults", "<project-specific allow entry>"],
    "soft_deny": ["$defaults", "<project-specific soft deny entry>"],
    "environment": ["$defaults", "<project-specific environment entry>"]
  }
}
```

安全条件:

- `"$defaults"` を削らない。
- `"$defaults"` を「Harness が考える default 一覧」に展開して書き直さない。
- 既存の `permissions.deny` / `permissions.ask` / `sandbox.failIfUnavailable` / `sandbox.network.deniedDomains` / `sandbox.filesystem` は維持する。
- Auto Mode entry を追加する場合も、deny / ask / sandbox を緩める理由にはしない。

R05 guardrail and sandbox.network.deniedDomains are not duplicated by Auto Mode:

- R05 は Go guardrail 側の危険操作検出で、`sudo` wrapper、`find -delete` / `find -exec rm ...`、macOS dangerous removal path などをコマンド文字列から検出する。Auto Mode は自動承認の分類層であり、R05 のような Harness 固有の二層目ガードとは責務が違う。
- `sandbox.network.deniedDomains` は metadata endpoint への到達を sandbox の network 境界で止める。Auto Mode の `environment` guidance は「どの環境条件なら自動化しやすいか」の分類であり、network deny list の代替ではない。
- `permissions.deny` / `permissions.ask` は、明示的な拒否・確認ルールとして残す。Claude Code 側の deny precedence と Harness の guardrail を重ねることで、Auto Mode default が更新されても破壊的操作や機密読み取りの防御を緩めない。

検証:

- `tests/test-claude-upstream-integration.sh` で、この section、template note、`.claude-plugin/settings.json` の既存 deny / ask / deniedDomains 維持を固定する。
- 将来 `.claude-plugin/settings.json` に `autoMode` を追加する場合は、`allow` / `soft_deny` / `environment` の各 entry に `"$defaults"` が含まれることを同 test の jq check が要求する。

## 53.1.5 plugin / managed settings policy

対象:

- Claude Code `2.1.118` の plugin `themes/` directory
- `DISABLE_UPDATES` と既存 `DISABLE_AUTOUPDATER`
- Claude Code `2.1.117` の `blockedMarketplaces` / `strictKnownMarketplaces`
- plugin dependency auto-resolve / missing dependency hints
- Windows / WSL managed settings 継承 (`wslInheritsWindowsSettings`)

今回の判断:

- `docs/plugin-managed-settings-policy.md` を新設し、setup / plugin policy docs の正本として扱う。
- `skills/harness-setup/SKILL.md` から同 docs へ pointer を追加し、setup 時に marketplace policy と dependency policy を迷わないようにする。
- `DISABLE_AUTOUPDATER` は自動更新停止、`DISABLE_UPDATES` は手動 `claude update` まで止める企業管理向けの強い停止として区別する。
- `blockedMarketplaces` / `strictKnownMarketplaces` は managed settings 専用の管理環境向け policy として扱い、通常ユーザー向け default には過剰適用しない。
- 通常の team onboarding では `extraKnownMarketplaces` を優先し、strict allowlist が必要な企業だけが managed settings で `strictKnownMarketplaces` を使う。
- plugin dependency auto-resolve と missing dependency hints は Claude Code 本体に任せる。Harness 独自の dependency resolver、cache 直接編集、marketplace policy 迂回は追加しない。
- plugin `themes/` directory は今回は `P` に留める。Harness は運用安全性の plugin であり、theme 同梱には brand / accessibility / terminal compatibility の別レビューが必要なため推測実装しない。
- `wslInheritsWindowsSettings` は Windows / WSL 混在企業環境向けの managed settings 候補として記録し、Harness default には入れない。

安全条件:

- `.claude-plugin/settings.json` に `DISABLE_UPDATES`、`blockedMarketplaces`、`strictKnownMarketplaces` を default として追加しない。
- managed settings の最上位 precedence と Claude Code 本体の install / update / refresh / auto-update enforcement を信頼する。
- Harness は説明・release guidance・検証 grep に留め、信頼境界そのものを再実装しない。
- この guidance は `tests/test-claude-upstream-integration.sh` で、policy docs の存在、`DISABLE_UPDATES` / marketplace policy / dependency resolver / themes decision の記述、Feature Table の完了表記を固定する。

## 53.1.6 Claude Code UX automatic inheritance policy

対象:

- Claude Code `2.1.118` の `/cost` / `/stats` から `/usage` への統合
- Claude Code `2.1.118` の `/continue` / `/resume` が `/add-dir` で追加された current directory の session を見つける改善
- Claude Code `2.1.117` の main-thread `--agent` + agent frontmatter `mcpServers` 読み込み
- Claude Code `2.1.117` の `CLAUDE_CODE_FORK_SUBAGENT=1` external build flag
- Claude Code `2.1.117` の stale large session summary、native `bfs` / `ugrep` search、高 effort default

今回の判断:

- `/usage` を利用量・コスト・統計の primary entrypoint として扱う。
- `/cost` / `/stats` は legacy typing shortcut として扱う。古い docs でコスト確認や統計確認の入口を説明する場合は、まず `/usage` を案内し、必要な tab を開く shortcut として `/cost` / `/stats` を補足する。
- `/resume` が `/add-dir` session を見つける改善と stale large session summary は Claude Code 本体の session discovery / summary logic を自動継承する。Harness は duplicate resume index、独自 stale-session summarizer、transcript 再読 wrapper を追加しない。
- `--agent` + `mcpServers` は agents audit の後続候補に残す。main-thread agent で MCP 前提の agent frontmatter がどう読まれるかは、既存 agent definitions と MCP setup guidance の棚卸しが必要なため、今回の task では `P` として記録する。
- `CLAUDE_CODE_FORK_SUBAGENT=1` は Harness default に強制しない。external build で forked subagent を検証するための upstream flag として扱い、配布 plugin の settings / skill default / environment template には入れない。
- native `bfs` / `ugrep` search は wrapper を追加しない。検索の高速化は Claude Code native macOS / Linux build が Bash 経由で提供する領域で、Harness が別 search shim を重ねると path 解決、glob 差異、fallback 差異を増やす。
- 高 effort default は Claude Code 本体の model/account policy として自動継承する。Harness の `harness-work` effort scoring は、複雑な task に `ultrathink` や `high` を上乗せする局所 policy に留め、Pro / Max subscriber 向けの built-in default を固定値で上書きしない。

Harness wrapper を追加しない理由:

- これらは UI command routing、session discovery、agent frontmatter loading、native search、model/account default のように Claude Code 本体が runtime で判断する領域。
- Harness が同じ判断を wrapper として再実装すると、upstream 側の修正後も古い挙動を抱えたり、ユーザーの managed settings / account policy / platform-specific build と衝突したりする。
- Harness の責務は、古い説明を `/usage` 中心へ更新し、`C` は自動継承として記録し、agent audit が必要な `--agent` + `mcpServers` と external build flag は `P` として後続候補に残すこと。

検証:

- `tests/test-claude-upstream-integration.sh` で、この section、`/usage` primary entrypoint、legacy shortcut の扱い、`--agent` + `mcpServers` follow-up、external build flag を default にしない方針、native search / high effort default の自動継承、Feature Table の `C/P` 表記を固定する。

## 53.2.1 Codex provider and model metadata setup policy

対象:

- Codex `0.123.0` の built-in `amazon-bedrock` model provider
- `model_providers.amazon-bedrock.aws.profile`
- Codex `0.123.0` の bundled model metadata refresh と current `gpt-5.4` default
- 古い固定 model slug の setup guidance 残留
- Claude Code 側 Bedrock guidance との切り分け

今回の判断:

- `docs/codex-provider-setup-policy.md` を新設し、Codex provider / model metadata setup guidance の正本として扱う。
- Bedrock を使う user / project だけが `model_provider = "amazon-bedrock"` と `[model_providers.amazon-bedrock.aws] profile = "codex-bedrock"` を自分の Codex config に追加する。
- Harness の配布用 `codex/.codex/config.toml` には、`amazon-bedrock` の説明コメントだけを置く。実際の `model_provider` default は設定しない。
- Harness は AWS credential、temporary token、secret key、Bedrock endpoint override を書き込まない。
- `gpt-5.4` は Codex `0.123.0` の current bundled model metadata として扱う。Harness setup は `model = "gpt-5.4"` を default として固定しない。
- `scripts/check-codex.sh` の古い `gpt-5.2-codex` 推奨 sample は削除し、通常は Codex CLI の current default metadata に任せる説明へ変更する。
- Claude Code 側の Bedrock guidance は `CLAUDE_CODE_USE_BEDROCK`、`ANTHROPIC_DEFAULT_*`、`modelOverrides` の領域として残す。Codex の `model_provider = "amazon-bedrock"` と混ぜない。

Bedrock config example:

```toml
model_provider = "amazon-bedrock"

[model_providers.amazon-bedrock.aws]
profile = "codex-bedrock"
```

古い固定 model slug の点検:

```bash
rg -n "gpt-5\.2-codex|gpt-5-codex|gpt-5\.1|codex-mini|gpt-5\.3-codex|gpt-5\.4" \
  docs skills codex skills-codex scripts tests templates .claude-plugin opencode .agents -u
```

結果の扱い:

- `scripts/check-codex.sh` の `gpt-5.2-codex` sample は setup guidance として古いため削除する。
- `scripts/codex-loop.sh`、`scripts/config-utils.sh`、advisor contract tests の `gpt-5.4` は Advisor Strategy の model policy / fixture であり、Codex setup default ではないため今回は維持する。
- `docs/CLAUDE-feature-table.md` の過去 version 説明にある Bedrock / model 名は履歴説明として維持する。
- 新規 setup docs / skill / Codex README では、古い model slug を推奨値として追加しない。

検証:

- `tests/test-claude-upstream-integration.sh` で、provider policy docs、`harness-setup` pointer、Codex README / config note、古い `gpt-5.2-codex` sample 削除、Feature Table の 53.2.1 完了表記を grep 固定する。

## 53.2.2 Codex MCP diagnostics and plugin loading policy

対象:

- Codex `0.123.0` の `/mcp verbose`
- Codex `0.123.0` の plugin `.mcp.json` loading
- plugin `.mcp.json` の `mcpServers` 形式
- plugin `.mcp.json` の top-level server map 形式

今回の判断:

- `docs/codex-mcp-diagnostics.md` を新設し、Codex MCP diagnostics / plugin MCP loading guidance の正本として扱う。
- 普段の Codex TUI では `/mcp` を軽量な server 状態確認として使う。
- MCP server が見えない、起動エラーが分からない、resources / resource templates の有無を見たい時だけ `/mcp verbose` を使う。
- `/mcp verbose` は diagnostics、resources、resource templates を見る troubleshoot 用の入口として案内する。
- plugin 内 `.mcp.json` は `mcpServers` 形式と top-level server map 形式の両方を受け取れる前提に更新する。
- 新規 plugin では、他 tool と共有しやすい `mcpServers` 形式を優先する。
- 既存 plugin が top-level server map 形式なら、Codex 側の loading 改善を利用し、不要な migration を要求しない。

`.mcp.json` examples:

```json
{
  "mcpServers": {
    "docs": {
      "command": "node",
      "args": ["server.js"]
    }
  }
}
```

```json
{
  "docs": {
    "command": "node",
    "args": ["server.js"]
  }
}
```

Claude Code 側 MCP guidance と混ぜない理由:

- Codex TUI の `/mcp` / `/mcp verbose` は Codex runtime の診断入口。
- Codex plugin `.mcp.json` loading は Codex plugin 側の読み込み改善。
- Claude Code 側の `claude mcp ...`、`.claude/mcp.json`、hook `type: "mcp_tool"` は別 surface。
- 53.1.2 の `type: "mcp_tool"` hook safety decision は Claude Code hooks の話であり、53.2.2 の Codex `/mcp verbose` guidance とは責務を分ける。

検証:

- `tests/test-claude-upstream-integration.sh` で、`docs/codex-mcp-diagnostics.md`、`harness-setup` pointer、Codex README guidance、`/mcp verbose`、diagnostics / resources / resource templates、`mcpServers` 形式、top-level server map 形式、Claude Code 側 MCP guidance と混ぜない方針、Feature Table の 53.2.2 完了表記を grep 固定する。
- `tests/test-codex-package.sh` で、Codex README の `/mcp verbose` と `.mcp.json` loading guidance を検出する。

## 53.2.3 Codex realtime handoff silence policy

対象:

- Codex `0.123.0` の realtime handoff
- background agents が受け取る transcript delta
- `harness-loop` の background runner
- `breezing` の Worker / Advisor / Reviewer
- advisor / reviewer drift 検知

今回の判断:

- Codex `0.123.0` の realtime handoff 改善は `A: docs / guidance 化済み` として取り込む。
- background agent が transcript delta を受け取れることは、途中通知を増やす理由ではなく、必要な時だけ判断を更新できる前提として扱う。
- `skills-codex/harness-loop/SKILL.md` と Codex mirror `codex/.codex/skills/harness-loop/SKILL.md` に `Realtime Handoff / Silence Policy` を追加する。
- `skills-codex/breezing/SKILL.md` と Codex mirror `codex/.codex/skills/breezing/SKILL.md` に、Worker / Advisor / Reviewer の silence policy を追加する。
- 共有 `skills/breezing/SKILL.md` と `skills/harness-loop/SKILL.md` には、長時間実行時の通知整理として同じ考え方を反映する。
- `scripts/codex-loop.sh` が生成する 1-cycle prompt に、transcript delta だけで余計な途中報告を出さない指示を追加する。

silence policy:

- 報告するのは cycle / task 完了、blocked、validation failure、review `REQUEST_CHANGES`、advisor `STOP`、plateau、contract readiness failure、user が明示的に status を求めた時。
- `advisor-request.v1` 未応答、`review-result.v1` 未到着、review loop plateau などの advisor / reviewer drift は silence 対象にしない。
- transcript delta を受け取っただけで task status、review verdict、advisor decision が変わっていない場合は明示的に沈黙する。
- tool stdout の細かな増分は log / status 側に寄せる。
- default は `harness-loop` では「1 cycle につき最終報告 1 回」、`breezing` では「task 完了ごとに progress feed 1 回」。

advisor / reviewer drift と矛盾しない理由:

- silence policy は「不要な通知を減らす」ための方針であり、品質判定や停止条件を弱めるものではない。
- Advisor は `PLAN` / `CORRECTION` / `STOP` の相談役、Reviewer は `APPROVE` / `REQUEST_CHANGES` の品質判定役として分離したままにする。
- drift は `.claude/state/session.events.jsonl` / contract / review artifact の欠落として扱い、会話上の沈黙とは別の異常として検出する。

検証:

- `tests/test-claude-upstream-integration.sh` で、snapshot、Codex harness-loop / breezing、共有 harness-loop / breezing、`scripts/codex-loop.sh` に silence policy と drift 例外があることを grep 固定する。
- `tests/test-codex-package.sh` で、Codex README と Codex skill mirror に realtime handoff / silence policy があることを検出する。
- `./scripts/sync-skill-mirrors.sh --check` で `skills-codex` と Codex mirror の drift がないことを確認する。

## Harness judgement

53.1.1 では snapshot を作るだけに留め、後続 task の実装を先取りしない。

実装候補として確定したもの:

- `type: "mcp_tool"` hook: 53.1.2 で no-op + safety test として完了。manifest 追加は必須 field 仕様と常設 read-only diagnostic tool が揃った後に行う
- `claude plugin tag`: 53.1.3
- Auto Mode `"$defaults"`: 53.1.4
- plugin themes / managed settings / update controls / dependency auto-resolve docs: 53.1.5
- Claude Code UX の自動継承 / 将来候補整理: 53.1.6
- Codex Bedrock provider / model metadata: 53.2.1
- Codex `/mcp verbose` / `.mcp.json` loading: 53.2.2
- Codex realtime handoff silence policy: 53.2.3
- Codex `remote_sandbox_config` / `codex exec` shared flags: 53.2.4
- Codex automatic bug fix notes: 53.2.5

自動継承に留める理由:

- UI / TUI / keyboard / OAuth / file watcher / Remote Control / runtime stability fix は、Claude Code または Codex 本体が直す領域で、Harness が wrapper を足すと挙動差分や二重責務を作りやすい。
- `/resume`、`/fork`、subagent cwd restore、Codex follow-up queue などは長時間作業に効くが、まず本体挙動を受け取るのが安全。Harness 側は必要な docs/guidance だけ後続 task で更新する。
- Codex `0.123.0` の app-server protocol / namespaced dynamic tools は release body にある範囲だけ記録し、compare から推測実装しない。

## Why `B: 書いただけ` is 0

この snapshot では `B: 書いただけ` を分類として使わない。

- `A` は必ず Phase 53 の具体的な Plans task に接続した。Feature Table に書くだけでは終わらせない。
- `C` は本体修正を自動継承する理由と、Harness 側で wrapper を追加しない理由を明記した。
- `P` は「今回は実装しない」ことを明示し、推測実装を避けた。

そのため、`CHANGELOG.md` と `docs/CLAUDE-feature-table.md` は snapshot への入口であり、一次情報と判断根拠の正本はこの文書に置く。
