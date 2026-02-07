import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/ttl-cache");
  const CacheClass = (module as any).TTLCache ?? (module as any).default;
  if (!CacheClass) throw new Error("No TTLCache export found");
  await runTests(() => new CacheClass(200));
});

async function runTests(create: () => any) {
  // basic set/get
  {
    const c = create();
    c.set('a', 'val');
    expect(c.get('a')).toBe('val');
  }
  // get expired
  {
    const c = create();
    c.set('a', 'val', 50);
    await new Promise(r => setTimeout(r, 80));
    expect(c.get('a')).toBeUndefined();
  }
  // has
  {
    const c = create();
    c.set('a', 1);
    expect(c.has('a')).toBe(true);
    expect(c.has('b')).toBe(false);
  }
  // delete
  {
    const c = create();
    c.set('a', 1);
    expect(c.delete('a')).toBe(true);
    expect(c.get('a')).toBeUndefined();
  }
  // clear
  {
    const c = create();
    c.set('a', 1);
    c.set('b', 2);
    c.clear();
    expect(c.get('a')).toBeUndefined();
  }
  // getOrSet new key
  {
    const c = create();
    const v = await c.getOrSet('x', () => 42);
    expect(v).toBe(42);
    expect(c.get('x')).toBe(42);
  }
  // getOrSet existing key
  {
    const c = create();
    c.set('x', 'old');
    const v = await c.getOrSet('x', () => 'new');
    expect(v).toBe('old');
  }
  // getOrSet async factory
  {
    const c = create();
    const v = await c.getOrSet('y', async () => {
      await new Promise(r => setTimeout(r, 10));
      return 'async';
    });
    expect(v).toBe('async');
  }
  // size accuracy with expiry — size() must not count expired entries
  {
    const c = create();
    c.set('a', 1, 50);
    c.set('b', 2, 5000);
    await new Promise(r => setTimeout(r, 80));
    expect(c.has('a')).toBe(false);
    expect(c.has('b')).toBe(true);
    // size() must reflect actual live entries only
    expect(c.size()).toBe(1);
  }
}

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", { encoding: "utf-8", stdio: "pipe" });
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
