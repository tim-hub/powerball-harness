# テストスイート

このディレクトリには、claude-code-harnessプラグインの品質を保証するためのテストが含まれています。

## VibeCoder向けテスト

エンタープライズレベルの複雑なテストではなく、**1人でクライアントプロジェクトをこなすVibeCoder**が、プラグインが正しく動作することを簡単に確認できるシンプルなテストです。

## テストの実行方法

### プラグイン構造の検証

プラグインの基本構造が正しいかを検証します：

```bash
./tests/validate-plugin.sh
./tests/validate-plugin-v3.sh
./scripts/ci/check-consistency.sh
```

### Unified Memory 検証

共通メモリdaemonの基本動作を検証します：

```bash
./tests/test-memory-daemon.sh
```

ゾンビプロセスが残らないかをループ検証します：

```bash
./tests/test-memory-daemon-zombie.sh 100
```

検索品質（hybrid ranking / privacy filter / API経路）を検証します：

```bash
./tests/test-memory-search-quality.sh
```

これらの検証は以下を確認します：

1. **プラグイン構造**: plugin.jsonの存在と妥当性
2. **コマンド**: 登録されたコマンドファイルの存在
3. **スキル**: スキル定義の存在と基本的な品質
4. **エージェント**: エージェント定義の存在
5. **フック**: hooks.jsonの妥当性
6. **スクリプト**: 自動化スクリプトの存在と実行権限
7. **ドキュメント**: README等の必須ドキュメント

### 期待される出力

```
==========================================
Claude harness - プラグイン検証テスト
==========================================

1. プラグイン構造の検証
----------------------------------------
✓ plugin.json が存在します
✓ plugin.json は有効なJSONです
✓ plugin.json に name フィールドがあります
✓ plugin.json に version フィールドがあります
...

==========================================
テスト結果サマリー
==========================================
合格: 25
警告: 1
失敗: 0

✓ 全てのテストに合格しました！
```

## テストの追加

新しいコマンドやスキルを追加した場合、このテストを実行して構造が正しいことを確認してください。

## CI/CDでの利用

GitHub Actions では `.github/workflows/validate-plugin.yml` が以下を実行します。

- `./tests/validate-plugin.sh`
- `./scripts/ci/check-consistency.sh`
- `./tests/test-codex-package.sh`
- `cd core && npm test`

`/harness-work all` の success / failure fixture は smoke / full を分けて管理しています。詳細は [docs/evidence/work-all.md](../docs/evidence/work-all.md) を参照してください。

## トラブルシューティング

### jqコマンドが見つからない

テストスクリプトは`jq`コマンドを使用します。インストールされていない場合：

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Windows (WSL)
sudo apt-get install jq
```

### テストが失敗する場合

1. エラーメッセージを確認
2. 該当するファイルが存在するか確認
3. JSONファイルの構文エラーがないか確認

## VibeCoder向けのポイント

- **シンプル**: 複雑なテストフレームワークは不要
- **実用的**: 実際に問題になる構造エラーを検出
- **高速**: 数秒で完了
- **わかりやすい**: 結果が一目でわかる

このテストは、プラグインを変更した後に「壊れていないか」を素早く確認するためのものです。
