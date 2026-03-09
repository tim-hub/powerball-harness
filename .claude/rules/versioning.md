# バージョニングルール

Harness のバージョン管理基準。SemVer（Semantic Versioning）に準拠する。

## バージョン判定基準

| 変更の種類 | バージョン | 例 |
|-----------|----------|-----|
| スキル定義（SKILL.md）の文言修正・追記 | **patch** (x.y.Z) | テンプレート微修正、説明文改善 |
| ドキュメント・ルールファイルの更新 | **patch** (x.y.Z) | CHANGELOG 書き換え、rules/ 追加 |
| hooks/scripts のバグ修正 | **patch** (x.y.Z) | task-completed.sh のエスケープ修正 |
| 既存スキルに新しいフラグ/サブコマンド追加 | **minor** (x.Y.0) | `--snapshot`、`--auto-mode` |
| 新しいスキル/エージェント/hooks 追加 | **minor** (x.Y.0) | 新スキル `harness-foo` |
| TypeScript ガードレールエンジンの変更 | **minor** (x.Y.0) | 新ルール追加、既存ルール変更 |
| Claude Code 新バージョン互換対応 | **minor** (x.Y.0) | CC v2.1.72 対応 |
| 破壊的変更（旧スキル廃止、フォーマット非互換） | **major** (X.0.0) | Plans.md v1 サポート削除 |

## 判断フローチャート

```
既存の動作が壊れる？
├─ Yes → major
└─ No → ユーザーが新しいことをできるようになる？
    ├─ Yes → minor
    └─ No → patch
```

## バッチリリースの推奨

- **同日に複数 Phase を完了した場合**: 1つの minor リリースにまとめる
- **Phase の完了 + ドキュメント修正**: Phase 分を minor、ドキュメント修正は同梱（別リリースにしない）
- **CC 互換対応 + 機能追加**: 1つの minor にまとめてよい

### 悪い例

```
v3.6.0 (03/08 AM) — Phase 25
v3.7.0 (03/08 PM) — Phase 26    ← 同日に 2 minor は避ける
v3.7.1 (03/09)    — Auto Mode
```

### 良い例

```
v3.6.0 (03/08) — Phase 25 + Phase 26    ← まとめて 1 minor
v3.6.1 (03/09) — Auto Mode 準備         ← prep は patch
```

## リリース前チェック

1. **前回リリースからの変更を一覧化**
2. **判定基準に照らしてバージョン種別を決定**
3. **同日の複数変更はバッチ化を検討**
4. **VERSION / plugin.json / CHANGELOG の3点同期を確認**
5. **git tag が欠番なく連続していることを確認**

## 禁止事項

- タグの削除・巻き戻し（公開済みバージョンは不変）
- 同日に 2 回以上の minor バンプ
- patch レベルの変更での minor バンプ
