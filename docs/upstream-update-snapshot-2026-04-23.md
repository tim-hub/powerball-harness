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
