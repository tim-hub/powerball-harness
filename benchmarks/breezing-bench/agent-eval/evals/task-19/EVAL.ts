import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/stack");
  const StackClass = (module as any).Stack ?? (module as any).default;
  if (!StackClass) throw new Error("No Stack export found");
  await runTests(() => new StackClass());
});

async function runTests(create: () => any) {
  // push and size
  {
    const s = create();
    s.push(1);
    s.push(2);
    s.push(3);
    expect(s.size()).toBe(3);
  }
  // pop LIFO order
  {
    const s = create();
    s.push('a');
    s.push('b');
    s.push('c');
    expect(s.pop()).toBe('c');
    expect(s.pop()).toBe('b');
    expect(s.pop()).toBe('a');
  }
  // pop empty
  {
    const s = create();
    expect(s.pop()).toBeUndefined();
  }
  // peek
  {
    const s = create();
    s.push(42);
    expect(s.peek()).toBe(42);
    expect(s.size()).toBe(1);
  }
  // peek empty
  {
    const s = create();
    expect(s.peek()).toBeUndefined();
  }
  // isEmpty
  {
    const s = create();
    expect(s.isEmpty()).toBe(true);
    s.push(1);
    expect(s.isEmpty()).toBe(false);
  }
  // toArray
  {
    const s = create();
    s.push(1);
    s.push(2);
    s.push(3);
    const arr = s.toArray();
    expect(arr).toHaveLength(3);
    expect(arr).toContain(1);
    expect(arr).toContain(2);
    expect(arr).toContain(3);
  }
  // toArray does not modify stack
  {
    const s = create();
    s.push('x');
    s.toArray();
    expect(s.size()).toBe(1);
  }
  // clear
  {
    const s = create();
    s.push(1);
    s.push(2);
    s.clear();
    expect(s.size()).toBe(0);
    expect(s.isEmpty()).toBe(true);
    expect(s.pop()).toBeUndefined();
  }
  // generic types
  {
    const s = create();
    s.push({ name: 'test' });
    expect(s.peek().name).toBe('test');
  }
}

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", { encoding: "utf-8", stdio: "pipe" });
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
