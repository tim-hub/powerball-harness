---
name: codex-worker
description: "Codexを下請けに。並列で実装を進めてもらう職人気質。Use when user mentions 'Codex に実装させて', 'Codex Worker', 'Codex に作らせて', or '実装を依頼'. Do NOT load for: 'Codex レビュー', 'セカンドオピニオン', 'Codex の意見', 'Codex でレビュー', 'Codex セットアップ'."
description-en: "Codex as subcontractor. Craftsman spirit in parallel implementation. Use when user mentions 'Codex に実装させて', 'Codex Worker', 'Codex に作らせて', or '実装を依頼'. Do NOT load for: 'Codex レビュー', 'セカンドオピニオン', 'Codex の意見', 'Codex でレビュー', 'Codex セットアップ'."
description-ja: "Codexを下請けに。並列で実装を進めてもらう職人気質。Use when user mentions 'Codex に実装させて', 'Codex Worker', 'Codex に作らせて', or '実装を依頼'. Do NOT load for: 'Codex レビュー', 'セカンドオピニオン', 'Codex の意見', 'Codex でレビュー', 'Codex セットアップ'."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep", "Task"]
argument-hint: "[task description]"
---

# Codex Worker Skill

Claude Code を PM/Orchestrator として、Codex を Worker として実装を委譲するスキル。

## Philosophy

> **「Claude Code = 設計・レビュー、Codex = 実装」**
>
> 高レベルな判断は Claude Code、実装の細部は Codex に任せる分業体制。

## Quick Reference

- "**Codex に実装させて**" → this skill
- "**Codex Worker**" → this skill
- "**Codex に作らせて**" → this skill
- "**実装を依頼**" → this skill

## Do NOT Load For (誤発動防止)

以下のキーワードは `codex-review` スキルが担当します（description と完全一致）:

| トリガーワード | 正しいスキル | 理由 |
|---------------|-------------|------|
| "**Codex レビュー**" | `codex-review` | レビュー = 品質チェック |
| "**セカンドオピニオン**" | `codex-review` | 意見取得 ≠ 実装 |
| "**Codex の意見**" | `codex-review` | 意見 ≠ Worker |
| "**Codex でレビュー**" | `codex-review` | レビュー目的 |
| "**Codex セットアップ**" | `codex-review` | MCP 設定 |

## Feature Details

| Feature | Reference |
|---------|-----------|
| **Setup** | See [references/setup.md](references/setup.md) |
| **Worker Execution** | See [references/worker-execution.md](references/worker-execution.md) |
| **Task Ownership** | See [references/task-ownership.md](references/task-ownership.md) |
| **Parallel Strategy** | See [references/parallel-strategy.md](references/parallel-strategy.md) |
| **Quality Gates** | See [references/quality-gates.md](references/quality-gates.md) |
| **Review & Integration** | See [references/review-integration.md](references/review-integration.md) |

## Execution Flow

```
1. タスク受信
    ↓
2. base-instructions 生成
   - Rules 連結
   - AGENTS.md 強制読み込み指示
    ↓
3. git worktree 準備（並列時）
    ↓
4. mcp__codex__codex 呼び出し
   - prompt: タスク内容 + AGENTS_SUMMARY 証跡出力指示
   - cwd: worktree パス
   - approval-policy: never
   - sandbox: workspace-write
    ↓
5. 結果検証
   - AGENTS_SUMMARY 証跡確認
   - 不合格時: 合計3回試行
    ↓
6. Orchestrator レビュー
   - 品質ゲート（lint, test, 改ざん検出）
   - 修正指示 → 再実行ループ
    ↓
7. マージ・Plans.md 更新
```

## MCP Parameters (D20)

```json
{
  "prompt": "タスク内容 + AGENTS_SUMMARY 証跡出力指示",
  "base-instructions": "Rules 連結 + AGENTS.md 強制読み込み指示",
  "cwd": "/path/to/worktree",
  "approval-policy": "never",
  "sandbox": "workspace-write"
}
```

## AGENTS.md Compliance

Worker は実行開始時に以下を出力する必要がある:

```
AGENTS_SUMMARY: <1行要約> | HASH:<SHA256先頭8文字>
```

- 入力: AGENTS.md ファイル内容（BOM除去、全行LF正規化）
- アルゴリズム: SHA256、Hex小文字、先頭8文字
- 欠落時: 即失敗 → 手動対応

## 完了時の必須アクション

**重要**: Codex の作業が完了したら、Claude Code は必ず以下を実行すること。

### Plans.md 更新手順

1. **該当タスクを特定**
   - ユーザーの依頼内容と Plans.md のタスクを照合
   - 該当するタスク行を見つける

2. **ステータス更新**
   ```markdown
   # Before
   - [ ] **タスク名** `cc:WIP`

   # After
   - [x] **タスク名** `cc:done`
   ```
   - チェックボックス: `[ ]` → `[x]`
   - マーカー: `cc:WIP` → `cc:done`

3. **更新しない場合**
   - Plans.md が存在しない
   - 該当するタスクが見つからない
   - ユーザーが明示的に「Plans.md は更新しないで」と指示した場合

### 例外処理

タスクが Plans.md にない場合は、ユーザーに確認:

```
✅ Codex による実装が完了しました。

Plans.md に該当タスクが見つかりませんでした。
- タスクを追加しますか？
- または既存タスクを更新しますか？
```

## MCP ツールアクセス（Claude Code 2.1.30+）

### Codex Worker での MCP ツール利用

Claude Code 2.1.30 以降、Codex Worker（MCP 経由で起動される Codex セッション）から SDK 提供 MCP ツールが共有可能になりました。

| MCP ツール | Codex Worker での利用 | 用途 |
|-----------|----------------------|------|
| **chrome-devtools** | ✅ 利用可能 | フロントエンド実装の動作確認 |
| **playwright** | ✅ 利用可能 | E2E テストの実装・検証 |
| **harness MCP** | ✅ 利用可能 | AST 検索、LSP 診断による品質向上 |
| **codex (再帰)** | ⚠️ 非推奨 | 無限ループのリスクあり |

### base-instructions での MCP ツール指示

Codex Worker に MCP ツールの活用を指示する例:

```markdown
## 利用可能なツール

以下の MCP ツールが利用可能です。実装品質向上のため活用してください:

- **harness_ast_search**: コードスメル検出（console.log 残留、空 catch 等）
- **harness_lsp_diagnostics**: 型エラー事前検出（ビルド前チェック）
- **playwright**: E2E テストが必要な場合の動作確認

### 推奨ワークフロー

1. 実装前: harness_lsp_diagnostics で既存コードの型情報確認
2. 実装中: エディタでの型チェックを意識
3. 実装後: harness_ast_search でコードスメル検出
4. テスト: 必要に応じて playwright で E2E 確認
```

### 並列 Codex Worker 実行時の注意

`/ultrawork --codex --parallel 3` のように複数 Worker が並列実行される場合:

| リソースタイプ | 注意点 | 対策 |
|--------------|--------|------|
| **Git worktree** | 各 Worker は独立した worktree で動作 | worktree ごとに MCP ツールの cwd を設定 |
| **ブラウザ** | chrome-devtools の同時アクセス | 順次実行またはポート分離 |
| **Codex API** | レート制限に注意 | 並列数を制限（推奨: 最大3並列） |
| **LSP サーバー** | 同一ファイルへの同時アクセス | タスク分割時にファイルを分離 |

### MCP ツール共有の仕組み

```
Claude Code (Orchestrator)
    ↓ MCP 設定を共有
    ├── Codex Worker #1 (worktree-1/)
    │     ├── harness_lsp_diagnostics → worktree-1/ で実行
    │     └── harness_ast_search → worktree-1/ で実行
    ├── Codex Worker #2 (worktree-2/)
    │     └── harness_lsp_diagnostics → worktree-2/ で実行
    └── Codex Worker #3 (worktree-3/)
          └── playwright → 独立ブラウザインスタンス
```

### 制限事項

| 制限 | 詳細 | 回避策 |
|------|------|-------|
| **cwd の扱い** | MCP 呼び出しに `cwd` パラメータを明示的に指定 | base-instructions で worktree パスを指示 |
| **承認ポリシー** | `approval-policy: never` 設定時は承認不要 | Orchestrator 側で設定済み |
| **Codex 再帰呼び出し** | Codex → Codex は無限ループのリスク | base-instructions で明示的に禁止 |

### トラブルシューティング

Codex Worker から MCP ツールが使えない場合:

1. **Claude Code のバージョン確認**
   ```bash
   claude --version
   # 2.1.30 以降であることを確認
   ```

2. **MCP サーバーの設定確認**
   - Orchestrator（Claude Code）側で MCP サーバーが設定されているか確認
   - `mcp__codex__codex` の `base-instructions` に MCP ツール利用指示が含まれているか確認

3. **フォールバック戦略**
   - MCP ツールが利用不可の場合、Codex は標準ツール（Read, Grep, Bash）にフォールバック
   - 品質は若干低下するが、実装は継続可能

---

## Related Skills

- `codex-review` - Codex によるレビュー・セカンドオピニオン
- `ultrawork` - `--codex` モードで Worker 並列実行
- `impl` - Claude Code 自身による実装
