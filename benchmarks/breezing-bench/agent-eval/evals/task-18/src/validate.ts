import { InvoiceCalculator } from './invoice-calculator.js';
import assert from 'node:assert';

const calc = new InvoiceCalculator();

// applyDiscount exists
assert(typeof calc.applyDiscount === 'function', 'applyDiscount must be implemented');

// Basic discount
const inv = calc.createInvoice('inv-1', [
  { description: 'Widget', quantity: 3, unitPrice: 111.00 }
], 0.10);

// 333.00 subtotal, 10% tax = 33.30, total = 366.30
const discounted = calc.applyDiscount('inv-1', 10);
assert(discounted, 'applyDiscount should return updated invoice');

// After 10% discount: subtotal 299.70, tax 29.97, total 329.67
const expectedSubtotal = 299.70;
const expectedTotal = 329.67;
assert(
  Math.abs(discounted!.subtotal - expectedSubtotal) < 0.01,
  `subtotal should be ~${expectedSubtotal}, got ${discounted!.subtotal}`
);
assert(
  Math.abs(discounted!.total - expectedTotal) < 0.01,
  `total should be ~${expectedTotal}, got ${discounted!.total}`
);

// Original invoice: precision check - total should be exactly 366.30 (rounded to 2dp)
const inv2 = calc.createInvoice('inv-2', [
  { description: 'Item', quantity: 3, unitPrice: 111.00 }
], 0.10);
const roundedTotal = Math.round(inv2.total * 100) / 100;
assert.strictEqual(roundedTotal, 366.30, 'total must be properly rounded to 2 decimal places');

console.log('All validations passed!');
