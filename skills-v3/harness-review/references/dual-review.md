# Dual Review (--dual)

Claude Reviewer と Codex Reviewer を並行実行し、異なるモデル視点でレビュー品質を向上させる。

## 前提条件

- Codex CLI がインストール済み（`scripts/codex-companion.sh setup --json` で確認）
- Codex が利用不可の場合、Claude 単独レビューにフォールバック

## 実行フロー

1. Codex の利用可否を確認する

   ```bash
   CODEX_AVAILABLE="$(bash scripts/codex-companion.sh setup --json 2>/dev/null | jq -r '.ready // false')"
   ```

2. Claude Reviewer を Task ツールで起動（通常の review フロー）

3. Codex が利用可能であれば `scripts/codex-companion.sh review` を並行起動

   ```bash
   # BASE_REF が渡されている場合は --base を指定。--json で構造化出力を取得
   bash scripts/codex-companion.sh review --base "${BASE_REF:-HEAD~1}" --json
   ```

4. 両方の結果を待ち合わせ

5. Verdict マージルール（以下の順に評価）:
   - 両方 APPROVE → `APPROVE`
   - どちらかが REQUEST_CHANGES → `REQUEST_CHANGES`（厳しい方を採用）
   - `critical_issues` は両方のリストを統合（重複排除なし）
   - `major_issues` は両方のリストを統合（重複排除なし）
   - `recommendations` は重複排除して統合

## 出力形式

通常の `review-result.v1` スキーマに `dual_review` フィールドを追加する:

```json
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVE | REQUEST_CHANGES",
  "dual_review": {
    "claude_verdict": "APPROVE | REQUEST_CHANGES",
    "codex_verdict": "APPROVE | REQUEST_CHANGES | unavailable | timeout",
    "merged_verdict": "APPROVE | REQUEST_CHANGES",
    "divergence_notes": "判定が分かれた場合の理由。例: Claude は Performance で major 検出、Codex は問題なし"
  },
  "critical_issues": [],
  "major_issues": [],
  "observations": [],
  "recommendations": []
}
```

### `codex_verdict` の特殊値

| 値 | 意味 |
|----|------|
| `"unavailable"` | Codex CLI がインストールされていないか利用不可 |
| `"timeout"` | Codex レビューがタイムアウト（120 秒以内に応答なし） |

## フォールバック

- **Codex が利用不可**: Claude 単独で実行し、`codex_verdict: "unavailable"` を記録する
- **Codex がタイムアウト**: Claude の verdict をそのまま採用し、`codex_verdict: "timeout"` を記録する
- **Codex のレビュー出力が不正**: パース失敗として扱い、`codex_verdict: "unavailable"` を記録する

いずれのフォールバックでも、Claude 単独レビューの結果が最終 verdict となる。

## Divergence Notes の書き方

判定が一致した場合（`claude_verdict == codex_verdict`）は `divergence_notes` を空文字列にする。

判定が分かれた場合は以下の形式で記録する:

```
Claude: REQUEST_CHANGES（Security - SQLインジェクションのリスク）
Codex: APPROVE（同箇所を問題なしと判定）
採用: REQUEST_CHANGES（厳しい方を優先）
```
