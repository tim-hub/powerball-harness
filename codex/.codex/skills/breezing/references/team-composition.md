# Team Composition

## Roles

- Lead: スコープ管理、進行、最終判定
- Implementers: 実装担当（並列）
- Reviewer: 品質判定担当

## Role Selection

| 条件 | Implementers | Reviewer |
|------|--------------|----------|
| デフォルト | Codex roles | Codex reviewer |
| `--claude` | Claude roles | Claude reviewer |

## Spawn Pattern

- `spawn_agent(agent_type="implementer")`
- `spawn_agent(agent_type="reviewer")`
- `wait(ids=[...])`

`--claude` 時は `agent_type` を `claude_implementer` / `claude_reviewer` に切り替える。
