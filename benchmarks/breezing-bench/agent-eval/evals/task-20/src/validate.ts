const mod = await import('./linked-list.js');
const ListClass = (mod as any).DoublyLinkedList ?? (mod as any).default;

import assert from 'node:assert';

assert(ListClass, 'DoublyLinkedList class must be exported');

const list = new ListClass();

// append + size
list.append(1);
list.append(2);
list.append(3);
assert.strictEqual(list.size(), 3, 'size should be 3');

// toArray
assert.deepStrictEqual(list.toArray(), [1, 2, 3], 'toArray should return [1,2,3]');

// prepend
list.prepend(0);
assert.deepStrictEqual(list.toArray(), [0, 1, 2, 3], 'prepend should add to front');

// find
const found = list.find((v: number) => v === 2);
assert.strictEqual(found, 2, 'find should return matching value');

const notFound = list.find((v: number) => v === 99);
assert.strictEqual(notFound, undefined, 'find should return undefined for no match');

// delete
assert.strictEqual(list.delete(2), true, 'delete existing should return true');
assert.deepStrictEqual(list.toArray(), [0, 1, 3], 'should remove deleted value');
assert.strictEqual(list.delete(99), false, 'delete non-existent should return false');

// reverse
list.reverse();
assert.deepStrictEqual(list.toArray(), [3, 1, 0], 'reverse should reverse the list');

console.log('All validations passed!');
