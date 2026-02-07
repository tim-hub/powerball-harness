import { EventEmitter } from './event-emitter.js';
import assert from 'node:assert';

const emitter = new EventEmitter();

// once: should fire exactly once
let count = 0;
emitter.once('ping', () => count++);
emitter.emit('ping');
emitter.emit('ping');
assert.strictEqual(count, 1, 'once listener should fire exactly once');

// off: removing listener A should not remove B
let callsA = 0, callsB = 0;
const listenerA = () => callsA++;
const listenerB = () => callsB++;
emitter.on('test', listenerA);
emitter.on('test', listenerB);
emitter.off('test', listenerA);
emitter.emit('test');
assert.strictEqual(callsA, 0, 'removed listener A should not be called');
assert.strictEqual(callsB, 1, 'remaining listener B should be called once');

console.log('All validations passed!');
