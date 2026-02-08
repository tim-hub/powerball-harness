---
name: memory
description: "Manage SSOT, memory, and cross-tool memory search. Guardian of decisions.md and patterns.md. Use when user mentions memory, SSOT, decisions.md, patterns.md, merging, migration, SSOT promotion, sync memory, save learnings, memory search, claude-mem, past decisions, record this, or cursor-mem integration. Do NOT load for: implementation work, reviews, ad-hoc notes, or in-session logging."
description-en: "Manage SSOT, memory, and cross-tool memory search. Guardian of decisions.md and patterns.md. Use when user mentions memory, SSOT, decisions.md, patterns.md, merging, migration, SSOT promotion, sync memory, save learnings, memory search, claude-mem, past decisions, record this, or cursor-mem integration. Do NOT load for: implementation work, reviews, ad-hoc notes, or in-session logging."
description-ja: "SSOTと記憶を管理し、ツール横断の記憶検索を提供。decisions.mdとpatterns.mdの守護者です。Use when user mentions memory, SSOT, decisions.md, patterns.md, merging, migration, SSOT promotion, sync memory, save learnings, memory search, claude-mem, past decisions, record this, or cursor-mem integration. Do NOT load for: implementation work, reviews, ad-hoc notes, or in-session logging."
allowed-tools: ["Read", "Write", "Edit", "Bash", "mcp__claude-mem__*"]
argument-hint: "[ssot|sync|migrate|search|record]"
context: fork
---

# Memory Skills

メモリとSSOT管理を担当するスキル群です。

## 機能詳細

| 機能 | 詳細 |
|------|------|
| **SSOT初期化** | See [references/ssot-initialization.md](references/ssot-initialization.md) |
| **Plans.mdマージ** | See [references/plans-merging.md](references/plans-merging.md) |
| **移行処理** | See [references/workflow-migration.md](references/workflow-migration.md) |
| **プロジェクト仕様同期** | See [references/sync-project-specs.md](references/sync-project-specs.md) |
| **メモリ→SSOT昇格** | See [references/sync-ssot-from-memory.md](references/sync-ssot-from-memory.md) |
| **記憶検索（Cursor連携）** | See [references/cursor-mem-search.md](references/cursor-mem-search.md) |

## 実行手順

1. ユーザーのリクエストを分類
2. 上記の「機能詳細」から適切な参照ファイルを読む
3. その内容に従って実行

## SSOT昇格

メモリシステム（Claude-mem / Serena）から重要な学びをSSOTに永続化します。

- "**Save what we learned**" → [references/sync-ssot-from-memory.md](references/sync-ssot-from-memory.md)
- "**Promote decisions to SSOT**" → [references/sync-ssot-from-memory.md](references/sync-ssot-from-memory.md)

## 記憶検索（Cursor連携）

CursorからClaude-memを活用し、セッション間の知識を引き継ぎます。

- "**過去の判断を確認したい**" → [references/cursor-mem-search.md](references/cursor-mem-search.md)
- "**この実装パターンを記録して**" → [references/cursor-mem-search.md](references/cursor-mem-search.md)
