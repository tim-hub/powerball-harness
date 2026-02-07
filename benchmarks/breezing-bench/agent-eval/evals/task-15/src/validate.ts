import { FormValidator } from './form-validator.js';
import assert from 'node:assert';

const v = new FormValidator();

// validateEmail basic
const emailResult = v.validateEmail('user@example.com');
assert.strictEqual(emailResult.valid, true, 'valid email should pass');

const badEmail = v.validateEmail('not-an-email');
assert.strictEqual(badEmail.valid, false, 'invalid email should fail');

// validateUrl basic
const urlResult = v.validateUrl('https://example.com');
assert.strictEqual(urlResult.valid, true, 'valid URL should pass');

const badUrl = v.validateUrl('not a url');
assert.strictEqual(badUrl.valid, false, 'invalid URL should fail');

// validate() with rules: valid form should be isValid=true
v.addRule({ field: 'name', required: true, minLength: 2 });
const result = v.validate({ name: 'Alice' });
assert.strictEqual(result.isValid, true, 'valid form data should return isValid=true');

const invalid = v.validate({ name: '' });
assert.strictEqual(invalid.isValid, false, 'empty required field should return isValid=false');

console.log('All validations passed!');
