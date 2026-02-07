import { parseHeaders } from './header-parser.js';
import assert from 'node:assert';

const mod = await import('./header-parser.js');
const parseSetCookie = (mod as any).parseSetCookie;
assert(typeof parseSetCookie === 'function', 'parseSetCookie must be exported');

// parseSetCookie basic
const cookie = parseSetCookie('session_id=abc123; Path=/; HttpOnly; Secure; SameSite=Lax');
assert.strictEqual(cookie.name, 'session_id');
assert.strictEqual(cookie.value, 'abc123');
assert.strictEqual(cookie.path, '/');
assert.strictEqual(cookie.httpOnly, true);
assert.strictEqual(cookie.secure, true);
assert.strictEqual(cookie.sameSite, 'Lax');

// parseSetCookie with max-age
const cookie2 = parseSetCookie('token=xyz; Max-Age=3600; Domain=.example.com');
assert.strictEqual(cookie2.maxAge, 3600);
assert.strictEqual(cookie2.domain, '.example.com');

// parseHeaders: values with colons must be preserved
const headers = parseHeaders('Authorization: Bearer token:secret\nContent-Type: text/html');
assert.strictEqual(
  headers['authorization'],
  'Bearer token:secret',
  'header values containing colons must be fully preserved'
);

console.log('All validations passed!');
