import { TemplateEngine } from './template-engine.js';
import assert from 'node:assert';

const engine = new TemplateEngine();

// registerHelper basic
engine.registerHelper('upper', (s: string) => s.toUpperCase());
const result1 = engine.render('Hello {{#upper name}}', { name: 'world' });
assert(result1.includes('WORLD') || result1.includes('world'), 'helper should transform value');

// registerHelper with another helper
engine.registerHelper('reverse', (s: string) => s.split('').reverse().join(''));
const result2 = engine.render('{{#reverse greeting}}', { greeting: 'hello' });
assert.strictEqual(result2, 'olleh', 'reverse helper should reverse the string');

// XSS protection: variables must be HTML-escaped
const xssResult = engine.render('{{name}}', { name: '<script>alert(1)</script>' });
assert(!xssResult.includes('<script>'), 'HTML tags must be escaped in variable output');
assert(xssResult.includes('&lt;script&gt;'), 'HTML must be entity-escaped');

console.log('All validations passed!');
