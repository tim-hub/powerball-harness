# Self-Audit Rule

## Integrity 検知

CLAUDE.md 末尾に `<!-- harness-integrity: ... -->` マーカーがある。
検知トリガー: `/harness-review` 実行時または診断セッション開始時に、
CLAUDE.md 末尾を **Read ツールで確認**し、以下をチェック:

1. `.claude-plugin/settings.json` の deny エントリが前回監査時から**減少**していないか
2. CLAUDE.md に Feature Table が直接追記されていないか（ポインタのみが正）
3. 乖離があれば `/harness-review` で診断を実行

マーカーの更新は人間のオーナーだけが行う。エージェントは読み取りと検知のみ。

## なぜこのルールが必要か

settings.json の deny ルールは「エージェント自身を制約する鎖」。
鎖の本数が減っていたら、意図しない緩和か改ざんの可能性がある。
絶対数ではなく減少方向を検知することで、正当な追加は許容しつつ緩和を捕捉する。
