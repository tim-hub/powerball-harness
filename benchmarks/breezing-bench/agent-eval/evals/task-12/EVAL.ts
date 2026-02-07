import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/priority-queue");
  const PQClass = (module as any).PriorityQueue ?? (module as any).default;
  if (!PQClass) throw new Error("No PriorityQueue export found");
  await runTests(() => new PQClass());
});

async function runTests(create: () => any) {
  // basic enqueue/dequeue
  {
    const q = create();
    q.enqueue('a', 2);
    q.enqueue('b', 1);
    q.enqueue('c', 3);
    expect(q.dequeue()).toBe('b');
    expect(q.dequeue()).toBe('a');
    expect(q.dequeue()).toBe('c');
  }
  // peek returns first without removing
  {
    const q = create();
    q.enqueue('x', 1);
    expect(q.peek()).toBe('x');
    expect(q.size()).toBe(1);
  }
  // peek on empty
  {
    const q = create();
    expect(q.peek()).toBeUndefined();
  }
  // priority 0 is valid
  {
    const q = create();
    q.enqueue('normal', 5);
    q.enqueue('urgent', 0);
    expect(q.peek()).toBe('urgent');
    expect(q.dequeue()).toBe('urgent');
  }
  // isEmpty
  {
    const q = create();
    expect(q.isEmpty()).toBe(true);
    q.enqueue('x', 1);
    expect(q.isEmpty()).toBe(false);
  }
  // size
  {
    const q = create();
    q.enqueue('a', 1);
    q.enqueue('b', 2);
    expect(q.size()).toBe(2);
    q.dequeue();
    expect(q.size()).toBe(1);
  }
  // dequeue empty
  {
    const q = create();
    expect(q.dequeue()).toBeUndefined();
  }
  // same priority preserves insertion order
  {
    const q = create();
    q.enqueue('first', 1);
    q.enqueue('second', 1);
    expect(q.dequeue()).toBe('first');
  }
}

test("typecheck passes", () => {
  let result = '';
  try {
    execSync("npx tsc --noEmit", { encoding: "utf-8", stdio: "pipe" });
  } catch (e: any) {
    result = e.stdout || e.stderr || '';
  }
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
