import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/event-emitter");
  const EmitterClass = (module as any).EventEmitter ?? (module as any).default;
  if (!EmitterClass) throw new Error("No EventEmitter export found");
  await runTests(() => new EmitterClass());
});

async function runTests(create: () => any) {
  // on + emit
  {
    const e = create();
    let val = 0;
    e.on('e', (n: number) => val += n);
    e.emit('e', 5);
    expect(val).toBe(5);
  }
  // off removes correct listener
  {
    const e = create();
    let a = 0, b = 0;
    const la = () => a++;
    const lb = () => b++;
    e.on('e', la);
    e.on('e', lb);
    e.off('e', la);
    e.emit('e');
    expect(a).toBe(0);
    expect(b).toBe(1);
  }
  // once fires once
  {
    const e = create();
    let c = 0;
    e.once('e', () => c++);
    e.emit('e');
    e.emit('e');
    expect(c).toBe(1);
  }
  // once passes args
  {
    const e = create();
    let v = '';
    e.once('e', (s: string) => v = s);
    e.emit('e', 'hello');
    expect(v).toBe('hello');
  }
  // listenerCount
  {
    const e = create();
    e.on('x', () => {});
    e.on('x', () => {});
    expect(e.listenerCount('x')).toBe(2);
    expect(e.listenerCount('y')).toBe(0);
  }
  // once auto-removes (listenerCount drops)
  {
    const e = create();
    e.once('x', () => {});
    expect(e.listenerCount('x')).toBe(1);
    e.emit('x');
    expect(e.listenerCount('x')).toBe(0);
  }
  // multiple once
  {
    const e = create();
    let a = 0, b = 0;
    e.once('e', () => a++);
    e.once('e', () => b++);
    e.emit('e');
    e.emit('e');
    expect(a).toBe(1);
    expect(b).toBe(1);
  }
}

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", { encoding: "utf-8", stdio: "pipe" });
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
