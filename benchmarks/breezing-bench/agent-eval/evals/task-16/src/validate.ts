import { ConfigMerger } from './config-merger.js';
import assert from 'node:assert';

const merger = new ConfigMerger();

// mergeWithStrategy exists
assert(typeof merger.mergeWithStrategy === 'function', 'mergeWithStrategy must be implemented');

// mergeWithStrategy replace
const result1 = merger.mergeWithStrategy(
  { tags: ['a', 'b'] } as any,
  { tags: ['c'] } as any,
  'replace'
);
assert.deepStrictEqual(result1.tags, ['c'], 'replace strategy should replace arrays');

// mergeWithStrategy append
const result2 = merger.mergeWithStrategy(
  { tags: ['a', 'b'] } as any,
  { tags: ['c'] } as any,
  'append'
);
assert.deepStrictEqual(result2.tags, ['a', 'b', 'c'], 'append strategy should concatenate arrays');

// merge must NOT mutate the base object
const defaults = { db: { host: 'localhost', port: 5432 } } as any;
const configA = { db: { host: 'server-a' } } as any;
const configB = { db: { host: 'server-b' } } as any;

merger.merge(defaults, configA);
const resultB = merger.merge(defaults, configB);

// After merging B, defaults.db.host should still be 'localhost' (not 'server-a')
assert.strictEqual(defaults.db.host, 'localhost', 'merge must not mutate the base object');

console.log('All validations passed!');
