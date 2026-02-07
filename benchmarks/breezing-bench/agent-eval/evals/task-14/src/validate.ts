import { TTLCache } from './ttl-cache.js';
import assert from 'node:assert';

const cache = new TTLCache<string>(100); // 100ms TTL

// getOrSet basic
const val = await cache.getOrSet('key1', () => 'hello');
assert.strictEqual(val, 'hello');
assert.strictEqual(cache.get('key1'), 'hello');

// getOrSet returns existing
cache.set('key2', 'existing');
const val2 = await cache.getOrSet('key2', () => 'new');
assert.strictEqual(val2, 'existing', 'getOrSet should return existing value');

// getOrSet with async factory
const val3 = await cache.getOrSet('key3', async () => {
  await new Promise(r => setTimeout(r, 10));
  return 'async-value';
});
assert.strictEqual(val3, 'async-value');

// size() must not count expired entries
cache.clear();
cache.set('expire-me', 'temp', 50);
cache.set('keep-me', 'permanent', 5000);
await new Promise(r => setTimeout(r, 80));
assert.strictEqual(cache.has('expire-me'), false, 'expired entry should not be found');
assert.strictEqual(cache.size(), 1, 'size() must only count live entries (not expired ones)');

console.log('All validations passed!');
