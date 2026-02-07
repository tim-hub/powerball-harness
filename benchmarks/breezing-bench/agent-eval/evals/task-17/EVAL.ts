import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/template-engine");
  const EngineClass = (module as any).TemplateEngine ?? (module as any).default;
  if (!EngineClass) throw new Error("No TemplateEngine export found");
  await runTests(() => new EngineClass());
});

async function runTests(create: () => any) {
  // basic variable substitution
  {
    const e = create();
    expect(e.render('Hello {{name}}!', { name: 'World' })).toBe('Hello World!');
  }
  // multiple variables
  {
    const e = create();
    expect(e.render('{{a}} and {{b}}', { a: 'X', b: 'Y' })).toBe('X and Y');
  }
  // missing variable
  {
    const e = create();
    expect(e.render('Hello {{name}}', {})).toBe('Hello ');
  }
  // HTML escaping - XSS prevention
  {
    const e = create();
    const result = e.render('{{content}}', { content: '<script>alert(1)</script>' });
    expect(result).not.toContain('<script>');
    expect(result).toContain('&lt;script&gt;');
  }
  // HTML escaping - quotes
  {
    const e = create();
    const result = e.render('{{val}}', { val: '"hello" & \'world\'' });
    expect(result).toContain('&amp;');
    expect(result).toContain('&quot;');
  }
  // registerHelper basic
  {
    const e = create();
    e.registerHelper('upper', (s: string) => s.toUpperCase());
    const result = e.render('{{#upper name}}', { name: 'hello' });
    expect(result).toBe('HELLO');
  }
  // registerHelper reverse
  {
    const e = create();
    e.registerHelper('reverse', (s: string) => s.split('').reverse().join(''));
    expect(e.render('{{#reverse word}}', { word: 'abc' })).toBe('cba');
  }
  // registerHelper with missing context key
  {
    const e = create();
    e.registerHelper('upper', (s: string) => s.toUpperCase());
    const result = e.render('{{#upper missing}}', {});
    expect(result).toBe('');
  }
  // multiple helpers
  {
    const e = create();
    e.registerHelper('upper', (s: string) => s.toUpperCase());
    e.registerHelper('lower', (s: string) => s.toLowerCase());
    expect(e.render('{{#upper a}} {{#lower b}}', { a: 'hi', b: 'BYE' })).toBe('HI bye');
  }
  // number context values
  {
    const e = create();
    expect(e.render('Count: {{n}}', { n: 42 })).toBe('Count: 42');
  }
}

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", { encoding: "utf-8", stdio: "pipe" });
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
