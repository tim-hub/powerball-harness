# Agent Frontmatter Policy

公式ドキュメント ([plugins-reference#agents](https://code.claude.com/docs/en/plugins-reference#agents)) が定義する
plugin agent の frontmatter 対応状況と、harness 各 agent への影響を記録する調査ドキュメント。

Phase 45.5.1 で調査・記録。実装変更は Phase 46 で行う。

---

## 1. 公式 plugin agent 対応 frontmatter 一覧

出典: https://code.claude.com/docs/en/plugins-reference#agents

### サポートされる field

| field | 型 / 有効値 | 説明 |
|-------|------------|------|
| `name` | string | agent の識別子 |
| `description` | string | agent の役割説明（auto-loading に使用） |
| `model` | string | 使用するモデル ID（例: `claude-sonnet-4-6`） |
| `effort` | `low` / `medium` / `high` | 推論強度。呼び出し側が指定（v2.1.68+） |
| `maxTurns` | number | 最大ターン数 |
| `tools` | string[] | 許可する Tool 名のリスト |
| `disallowedTools` | string[] | 禁止する Tool 名のリスト |
| `skills` | string[] | agent がロードするスキル名のリスト |
| `memory` | string | agent memory の scope（例: `project`） |
| `background` | boolean | バックグラウンド実行フラグ |
| `isolation` | `"worktree"` のみ valid | worktree 分離。それ以外の値は無効 |
| `initialPrompt` | string | agent セッション開始時の先頭プロンプト |

> **`initialPrompt` の状況**:
> 公式 plugins-reference には明示的な記載がない（2026-04-18 時点）。
> ただし CHANGELOG v2.1.83 エントリには「Agent `initialPrompt` frontmatter (v2.1.83)」として
> agents/ での安定化用途で記載されており、harness の Feature Table にも採用済み。
> plugin agent でサポートされるか否かは公式 docs では未確認。

### silently ignored field（security restriction）

| field | 理由 |
|-------|------|
| `hooks` | security restriction。プラグイン agent 内からのフック定義を防ぐため無視される |
| `mcpServers` | security restriction。同上 |
| `permissionMode` | security restriction。親セッションからの権限継承のみ許可 |

> これらは **silently ignored** — エラーにならず、frontmatter に書いても動作に影響しない。

---

## 2. harness 各 agent の frontmatter 監査表

各 agent の frontmatter を実際に確認し、公式対応状況を判定した。

| field | worker.md | reviewer.md | advisor.md | scaffolder.md | 公式対応 | 備考 |
|-------|-----------|-------------|-----------|---------------|---------|------|
| `name` | `worker` | `reviewer` | `advisor` | `scaffolder` | ✅ | |
| `description` | ✅ あり | ✅ あり | ✅ あり | ✅ あり | ✅ | |
| `tools` | Read, Write, Edit, Bash, Grep, Glob | Read, Grep, Glob | Read, Grep, Glob | Read, Write, Edit, Bash, Grep, Glob | ✅ | |
| `disallowedTools` | Agent | Write, Edit, Bash, Agent | Write, Edit, Bash, Agent | Agent | ✅ | |
| `model` | `claude-sonnet-4-6` | `claude-sonnet-4-6` | `claude-opus-4-6` | `claude-sonnet-4-6` | ✅ | |
| `effort` | `medium` | `xhigh` | `xhigh` | `medium` | ⚠️ | `xhigh` は v2.1.111 追加値。plugin agent での対応状況は未確認 |
| `maxTurns` | `100` | `50` | `20` | `75` | ✅ | |
| `permissionMode` | `bypassPermissions` | `bypassPermissions` | `bypassPermissions` | `bypassPermissions` | ❌ silently ignored | Worker の bypassPermissions が実は効いていない可能性がある |
| `color` | `yellow` | `blue` | `purple` | `green` | ⚠️ | 公式 docs 未記載。CC UI 表示用の拡張 field と思われる |
| `memory` | `project` | `project` | ❌ なし | `project` | ✅ | |
| `isolation` | `worktree` | ❌ なし | ❌ なし | ❌ なし | ✅ | `worktree` のみ valid |
| `initialPrompt` | ✅ あり | ✅ あり | ✅ あり | ✅ あり | ⚠️ | 公式 plugins-reference に明記なし。CHANGELOG v2.1.83 には記載あり |
| `skills` | harness-work, harness-review | harness-review | ❌ なし | harness-setup, harness-plan | ✅ | |
| `hooks` (worker.md) | PreToolUse: Write\|Edit matcher | ❌ なし | ❌ なし | ❌ なし | ❌ silently ignored | hook は発火していない可能性がある |
| `hooks` (reviewer.md) | ❌ なし | Stop hook あり | ❌ なし | ❌ なし | ❌ silently ignored | Reviewer の Stop hook も発火しない可能性 |

### 凡例

- ✅ 公式対応済み
- ❌ silently ignored（公式非対応）
- ⚠️ 公式 docs 未記載または要確認

---

## 3. silently ignored field の影響範囲分析

### permissionMode

worker.md / reviewer.md / advisor.md / scaffolder.md の全 agent が `permissionMode: bypassPermissions` を宣言しているが、
plugin agent では security restriction により **silently ignored** になる。
この場合、plugin agent が spawn される際は **親セッションの permission mode を継承する**。

breezing mode で Lead (親) が `bypassPermissions` で動作している場合、
spawn された Worker も同じ permission mode を継承するため、実運用上は影響軽微となることが多い。
ただし、isolated worktree 実行環境や Auto Mode の親セッションでは継承パスが不明確であり、
Worker が期待する「ファイル書き込み無制限」の権限を得られない可能性がある。
特に、Worker が制限された permission context 内で `git commit` や `Write` ツールを呼び出す場面では
Permission エラーが出る可能性があることを認識しておく必要がある。

### hooks

worker.md の frontmatter には `PreToolUse: matcher: "Write|Edit"` フックが定義されており、
これは `hooks/pre-tool.sh` を呼び出して書き込み系ツールの使用前にガードレール検証を行う想定だった。
しかし plugin agent では `hooks` が silently ignored になるため、**Worker spawn 時にこのフックは発火しない**。

同様に reviewer.md の `Stop` フックも発火しない。

現在、書き込み系フックの代替として plugin level の `hooks/hooks.json` が存在し、
`SubagentStart` / `PreToolUse` matcher で hook を仕掛けているが、
agent frontmatter の hooks との整合性が保たれているかを確認する必要がある。
agent 固有の hooks（例: Worker だけに適用したい guardrail）を plugin level の hooks.json で表現するには
`SubagentStart` の agent 名フィルタリングが必要になり、設計が複雑化する懸念がある。

### worker_worktree_share.md との関連

エージェントメモリ `worker_worktree_share.md`（Phase 44.6.1 / 44.7.1 で観測）に記録された
並列 Worker spawn での `isolation: worktree` が効かない現象は、
`isolation` フィールド自体は公式サポート済みである（❌ではなく ✅）ため、別の原因による。

ただし `permissionMode` が silently ignored になることで、
複数 Worker が異なる permission context を持てず親の context に依存するという間接的な影響がある。
並列 spawn 時に親セッション（Lead）のブランチ切り替えや stash 操作と競合する場合、
Worker が想定外の permission 状態で動作している可能性を排除できない。

---

## 4. 修正案（実装は次サイクル）

### 1. `permissionMode` を agent frontmatter から削除

worker.md / reviewer.md / advisor.md / scaffolder.md の frontmatter から `permissionMode: bypassPermissions` を削除する。
代わりに、breezing skill 側でタスク委託の `Agent()` 呼び出し時に親セッションの permission 設定を継承させる形を明文化する。
これにより frontmatter の「書いてあるが動かない設定」を排除し、実際の動作との乖離をなくす。

### 2. `hooks` を plugin level `hooks/hooks.json` に SubagentStart matcher 付きで移植

worker.md の `PreToolUse: Write|Edit` フックを `hooks/hooks.json` に移植する。
移植時は `SubagentStart` イベントの後に agent 名（worker）をフィルタリングして
Worker spawn 時にのみ書き込みガードを適用する設計とする。
ただし発火タイミングが agent frontmatter の hooks と微妙に異なる可能性（Worker の preflight との差）があるため、
実機検証（実際に breezing で hook が発火するか）を先行して行うことが推奨される。

### 3. `initialPrompt` の plugin agent サポート状況を CC docs/changelog で再確認

`initialPrompt` は CHANGELOG v2.1.83 には記載があるが、公式 plugins-reference には明示がない。
`agents/` 内の各エージェント定義では `initialPrompt` を活用して初期動作を安定化させているため、
もし plugin agent でサポートされていないことが判明した場合は、`initialPrompt` の内容を
agent 本文の先頭セクション（例: `## 開始時の確認手順`）に移植する必要がある。
次サイクルでは CC 公式 docs の更新確認または実機テスト（`initialPrompt` あり/なしでの動作比較）を実施すること。

---

## 5. 本 Phase で実装しない rationale

本 Phase（45.5.1）の実施時点で v4.2.0-arcana のリリースブランチが既にカットされており、
agent frontmatter の修正はリリーススコープ外の変更を意味する。
`permissionMode` や `hooks` の修正は agents/*.md を変更する必要があり、
breezing フローの動作に直接影響するため、修正方針の確定には実機検証が必要だ。
「hook が実際に発火しているか」「permissionMode なしで Worker が期待通りの権限で動作するか」という
実測なしに変更を加えることは、既存の稼働フローに予期しない副作用をもたらすリスクがある。
本 Phase では調査結果と修正案を記録し、Phase 46 以降のサイクルで実機検証と実装を行うことを推奨する。

---

## 関連ファイル

- `agents/worker.md` — Worker agent 定義
- `agents/reviewer.md` — Reviewer agent 定義
- `agents/advisor.md` — Advisor agent 定義
- `agents/scaffolder.md` — Scaffolder agent 定義
- `hooks/hooks.json` — plugin level フック定義
- `docs/CLAUDE-feature-table.md` — `initialPrompt` v2.1.83 エントリ
- `.claude/agent-memory/claude-code-harness-worker/worker_worktree_share.md` — 並列 Worker worktree 競合の観測記録
