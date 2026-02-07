HTTP ヘッダーパーサーに `parseSetCookie(header: string)` 関数を追加してください。
`types.ts` に型定義があります。
Set-Cookie ヘッダー文字列を解析し、name, value, 属性(expires, max-age, path, domain, secure, httponly, samesite)を返す機能です。
