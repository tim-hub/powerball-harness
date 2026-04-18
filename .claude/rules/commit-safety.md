# Commit Safety Rules

Harness でのコミット操作に関する安全運用ルール。
エージェントが意図せず commit を戻したり上書きしたりするリスクを防ぐ。

## /undo — セッション内の変更取り消し（CC 2.1.108+）

CC 2.1.108 で `/rewind` のエイリアスとして `/undo` が追加された。

### 動作定義

`/undo` は Claude Code のセッション内で直前のアクション（ツール呼び出し・ファイル変更）を
取り消す。**git commit の revert/reset とは異なる**。

| 操作 | 対象 | git への影響 |
|------|------|------------|
| `/undo` | CC セッション内の直前ツール呼び出し | 変更をディスクから元に戻すが、`git commit` 済みのものは戻らない |
| `git revert` | git コミット単位 | 新たな revert commit を作成 |
| `git reset --hard` | git コミット単位 | 不可逆。Harness deny ルールで保護 |

### Harness エージェントの利用制約

**Worker / Reviewer は `/undo` を自律的に実行しない。**

以下の条件を全て満たした場合のみ、Lead (ユーザー) の明示指示を受けて実行を許可する:

1. ユーザーが「直前の変更を取り消して」と明示した
2. 取り消し対象が git commit 前のファイル変更である（commit 済みは git revert を使う）
3. 影響ファイルが 1 セッション内での変更に限定される

### 禁止パターン

- `REQUEST_CHANGES` 対応で `/undo` を使って commit を消す行為（`git revert` を使うこと）
- Reviewer が「この変更は不要」と判断して自律的に `/undo` する行為
- 修正ループ中に amend の代わりに `/undo` を使う行為（`git commit --amend` を使うこと）

### /undo が有効な用途（参考）

- エージェントが誤ってファイルを上書きした直後に人間が取り消す
- セッション内のドライランで意図しないファイル書き込みを取り消す

### 関連ルール

- `git reset --hard` は `.claude-plugin/settings.json` の deny と guardrail R11 で保護済み
- `git push --force` は guardrail R06 と deny で保護済み
- 不可逆な git 操作はユーザー手動実行を要求すること（Permission Boundaries 参照）

詳細: [CLAUDE.md — Permission Boundaries](../../CLAUDE.md)
