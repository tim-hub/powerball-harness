import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/config-merger");
  const MergerClass = (module as any).ConfigMerger ?? (module as any).default;
  if (!MergerClass) throw new Error("No ConfigMerger export found");
  await runTests(() => new MergerClass());
});

async function runTests(create: () => any) {
  // basic merge
  {
    const m = create();
    const result = m.merge({ a: 1, b: 2 }, { b: 3, c: 4 });
    expect(result.a).toBe(1);
    expect(result.b).toBe(3);
    expect(result.c).toBe(4);
  }
  // deep merge
  {
    const m = create();
    const result = m.merge(
      { db: { host: 'localhost', port: 5432 } },
      { db: { host: 'prod-server' } }
    );
    expect(result.db.host).toBe('prod-server');
    expect((result.db as any).port).toBe(5432);
  }
  // merge must NOT mutate base
  {
    const m = create();
    const base = { db: { host: 'localhost', port: 5432 } };
    const overrideA = { db: { host: 'server-a' } };
    const overrideB = { db: { host: 'server-b' } };
    m.merge(base, overrideA);
    const resultB = m.merge(base, overrideB);
    expect(base.db.host).toBe('localhost');
    expect(resultB.db.host).toBe('server-b');
  }
  // mergeWithStrategy replace
  {
    const m = create();
    const result = m.mergeWithStrategy(
      { tags: ['a', 'b'] },
      { tags: ['c'] },
      'replace'
    );
    expect(result.tags).toEqual(['c']);
  }
  // mergeWithStrategy append
  {
    const m = create();
    const result = m.mergeWithStrategy(
      { tags: ['a', 'b'] },
      { tags: ['c'] },
      'append'
    );
    expect(result.tags).toEqual(['a', 'b', 'c']);
  }
  // mergeWithStrategy prefer-base
  {
    const m = create();
    const result = m.mergeWithStrategy(
      { name: 'base', extra: 'keep' },
      { name: 'override', added: 'new' },
      'prefer-base'
    );
    expect(result.name).toBe('base');
    expect(result.extra).toBe('keep');
    expect(result.added).toBe('new');
  }
  // mergeWithStrategy does not mutate
  {
    const m = create();
    const base = { items: ['x'] };
    m.mergeWithStrategy(base, { items: ['y'] }, 'append');
    expect(base.items).toEqual(['x']);
  }
}

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", { encoding: "utf-8", stdio: "pipe" });
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
