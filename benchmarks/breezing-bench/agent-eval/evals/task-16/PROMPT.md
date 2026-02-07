ConfigMerger に `mergeWithStrategy(base, override, strategy)` メソッドを追加してください。
`types.ts` にインターフェース定義があります。
strategy は 'replace' | 'append' | 'prefer-base' を指定でき、ネストオブジェクトの配列マージ戦略を制御します。
