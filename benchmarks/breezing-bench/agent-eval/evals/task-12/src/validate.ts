import { PriorityQueue } from './priority-queue.js';
import assert from 'node:assert';

const pq = new PriorityQueue<string>();

// peek: should return undefined for empty queue
assert.strictEqual(pq.peek(), undefined, 'peek on empty queue should return undefined');

// enqueue + peek
pq.enqueue('low', 10);
pq.enqueue('high', 1);
assert.strictEqual(pq.peek(), 'high', 'peek should return highest priority item');
assert.strictEqual(pq.size(), 2, 'peek should not remove item');

// priority 0 should be valid (highest priority)
const pq2 = new PriorityQueue<string>();
pq2.enqueue('normal', 5);
pq2.enqueue('urgent', 0);
assert.strictEqual(pq2.peek(), 'urgent', 'priority 0 must be valid and highest');

console.log('All validations passed!');
