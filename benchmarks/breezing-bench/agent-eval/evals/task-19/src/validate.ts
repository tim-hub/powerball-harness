// Dynamic import to handle any export name
const mod = await import('./stack.js');
const StackClass = (mod as any).Stack ?? (mod as any).default;

import assert from 'node:assert';

assert(StackClass, 'Stack class must be exported');

const stack = new StackClass();

// push + size
stack.push(1);
stack.push(2);
stack.push(3);
assert.strictEqual(stack.size(), 3, 'size should be 3 after 3 pushes');

// peek
assert.strictEqual(stack.peek(), 3, 'peek should return top item');
assert.strictEqual(stack.size(), 3, 'peek should not remove item');

// pop
assert.strictEqual(stack.pop(), 3, 'pop should return top item');
assert.strictEqual(stack.size(), 2, 'size should decrease after pop');

// isEmpty
assert.strictEqual(stack.isEmpty(), false, 'should not be empty');
stack.pop();
stack.pop();
assert.strictEqual(stack.isEmpty(), true, 'should be empty after popping all');

// pop on empty
assert.strictEqual(stack.pop(), undefined, 'pop on empty should return undefined');

// toArray
stack.push('a');
stack.push('b');
const arr = stack.toArray();
assert(Array.isArray(arr), 'toArray should return an array');
assert.strictEqual(arr.length, 2, 'toArray should have 2 items');

// clear
stack.clear();
assert.strictEqual(stack.size(), 0, 'clear should empty the stack');
assert.strictEqual(stack.isEmpty(), true, 'should be empty after clear');

console.log('All validations passed!');
