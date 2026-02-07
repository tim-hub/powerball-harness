import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/linked-list");
  const ListClass = (module as any).DoublyLinkedList ?? (module as any).default;
  if (!ListClass) throw new Error("No DoublyLinkedList export found");
  await runTests(() => new ListClass());
});

async function runTests(create: () => any) {
  // append + toArray
  {
    const l = create();
    l.append(1);
    l.append(2);
    l.append(3);
    expect(l.toArray()).toEqual([1, 2, 3]);
  }
  // prepend
  {
    const l = create();
    l.prepend(3);
    l.prepend(2);
    l.prepend(1);
    expect(l.toArray()).toEqual([1, 2, 3]);
  }
  // size
  {
    const l = create();
    expect(l.size()).toBe(0);
    l.append(1);
    expect(l.size()).toBe(1);
    l.append(2);
    expect(l.size()).toBe(2);
  }
  // delete head
  {
    const l = create();
    l.append(1);
    l.append(2);
    l.append(3);
    expect(l.delete(1)).toBe(true);
    expect(l.toArray()).toEqual([2, 3]);
  }
  // delete tail
  {
    const l = create();
    l.append(1);
    l.append(2);
    l.append(3);
    expect(l.delete(3)).toBe(true);
    expect(l.toArray()).toEqual([1, 2]);
  }
  // delete middle
  {
    const l = create();
    l.append(1);
    l.append(2);
    l.append(3);
    expect(l.delete(2)).toBe(true);
    expect(l.toArray()).toEqual([1, 3]);
  }
  // delete non-existent
  {
    const l = create();
    l.append(1);
    expect(l.delete(99)).toBe(false);
  }
  // delete only element
  {
    const l = create();
    l.append(1);
    expect(l.delete(1)).toBe(true);
    expect(l.size()).toBe(0);
    expect(l.toArray()).toEqual([]);
  }
  // find
  {
    const l = create();
    l.append('a');
    l.append('b');
    l.append('c');
    expect(l.find((v: string) => v === 'b')).toBe('b');
    expect(l.find((v: string) => v === 'z')).toBeUndefined();
  }
  // reverse
  {
    const l = create();
    l.append(1);
    l.append(2);
    l.append(3);
    l.reverse();
    expect(l.toArray()).toEqual([3, 2, 1]);
  }
  // reverse empty
  {
    const l = create();
    l.reverse();
    expect(l.toArray()).toEqual([]);
  }
  // reverse single
  {
    const l = create();
    l.append(1);
    l.reverse();
    expect(l.toArray()).toEqual([1]);
  }
  // mixed operations
  {
    const l = create();
    l.append(2);
    l.prepend(1);
    l.append(3);
    l.delete(2);
    l.reverse();
    expect(l.toArray()).toEqual([3, 1]);
    expect(l.size()).toBe(2);
  }
}

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", { encoding: "utf-8", stdio: "pipe" });
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
