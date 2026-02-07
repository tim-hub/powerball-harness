import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/form-validator");
  const ValidatorClass = (module as any).FormValidator ?? (module as any).default;
  if (!ValidatorClass) throw new Error("No FormValidator export found");
  await runTests(() => new ValidatorClass());
});

async function runTests(create: () => any) {
  // validateEmail valid
  {
    const v = create();
    expect(v.validateEmail('test@example.com').valid).toBe(true);
  }
  // validateEmail invalid
  {
    const v = create();
    expect(v.validateEmail('not-an-email').valid).toBe(false);
    expect(v.validateEmail('').valid).toBe(false);
    expect(v.validateEmail('@example.com').valid).toBe(false);
  }
  // validateUrl valid
  {
    const v = create();
    expect(v.validateUrl('https://example.com').valid).toBe(true);
    expect(v.validateUrl('http://example.com/path?q=1').valid).toBe(true);
  }
  // validateUrl invalid
  {
    const v = create();
    expect(v.validateUrl('not a url').valid).toBe(false);
    expect(v.validateUrl('').valid).toBe(false);
  }
  // validate isValid logic - valid data
  {
    const v = create();
    v.addRule({ field: 'name', required: true, minLength: 2 });
    const result = v.validate({ name: 'Alice' });
    expect(result.isValid).toBe(true);
    expect(Object.keys(result.errors)).toHaveLength(0);
  }
  // validate isValid logic - invalid data
  {
    const v = create();
    v.addRule({ field: 'name', required: true, minLength: 2 });
    const result = v.validate({ name: '' });
    expect(result.isValid).toBe(false);
    expect(result.errors['name']).toBeDefined();
  }
  // validate with no rules - should be valid
  {
    const v = create();
    const result = v.validate({ anything: 'value' });
    expect(result.isValid).toBe(true);
  }
  // validate multiple rules
  {
    const v = create();
    v.addRule({ field: 'email', required: true });
    v.addRule({ field: 'name', required: true, minLength: 3 });
    const result = v.validate({ email: 'a@b.com', name: 'Bo' });
    expect(result.isValid).toBe(false);
  }
  // validate maxLength
  {
    const v = create();
    v.addRule({ field: 'code', maxLength: 5 });
    expect(v.validate({ code: '12345' }).isValid).toBe(true);
    expect(v.validate({ code: '123456' }).isValid).toBe(false);
  }
}

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", { encoding: "utf-8", stdio: "pipe" });
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
