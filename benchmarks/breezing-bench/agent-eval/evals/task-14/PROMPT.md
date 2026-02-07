TTLCache に `getOrSet(key, factory, ttl?)` メソッドを追加してください。
`types.ts` にインターフェース定義があります。
キーが存在すればその値を返し、なければ factory 関数を呼んで結果をセットしてから返す機能です。
factory は非同期関数にも対応してください。
