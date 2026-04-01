# Release Preflight

`scripts/release-preflight.sh` は、公開前に「今 release してよいか」を先に止めるための read-only チェックです。
vendor-neutral を前提にしているので、AWS 固定や特定の deploy 基盤に依存しません。

## 何を見るか

- working tree が clean か
- `CHANGELOG.md` に `[Unreleased]` があるか
- `.env.example` と `.env` が大きくずれていないか。`.env` がない repo は warning に留め、managed secrets 前提の運用を止めすぎない
- 既存の `healthcheck` / `preflight` コマンドが通るか
- `agents/` / `core/` / `hooks/` / `scripts/` の shipped surface に `mockData` / `dummy` / `localhost` / `TODO` / `FIXME` などの残骸が残っていないかを警告する
- 取得できる場合は CI の最新状態が成功しているか

## 使い方

```bash
scripts/release-preflight.sh
scripts/release-preflight.sh --root /path/to/other/repo
```

## 環境変数

- `HARNESS_RELEASE_PROJECT_ROOT`: 別 repo を点検したいときの root
- `HARNESS_RELEASE_HEALTHCHECK_CMD`: repo 固有の healthcheck コマンド
- `HARNESS_RELEASE_CI_STATUS_CMD`: CI 状態確認を差し替えたいときのコマンド

## dry-run との関係

`/release --dry-run` でも preflight は必ず通す。
dry-run は「公開操作をしない」という意味で、preflight は「公開してよい状態かを確認する」という意味。
両者は別物なので、dry-run でも preflight は省略しない。
