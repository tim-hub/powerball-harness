import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/invoice-calculator");
  const CalcClass = (module as any).InvoiceCalculator ?? (module as any).default;
  if (!CalcClass) throw new Error("No InvoiceCalculator export found");
  await runTests(() => new CalcClass());
});

async function runTests(create: () => any) {
  // create invoice
  {
    const c = create();
    const inv = c.createInvoice('1', [{ description: 'A', quantity: 2, unitPrice: 10 }], 0.10);
    expect(inv.subtotal).toBe(20);
    expect(inv.tax).toBeCloseTo(2.0);
    expect(inv.total).toBeCloseTo(22.0);
  }
  // get invoice
  {
    const c = create();
    c.createInvoice('1', [], 0.10);
    expect(c.getInvoice('1')).toBeDefined();
    expect(c.getInvoice('2')).toBeUndefined();
  }
  // add item
  {
    const c = create();
    c.createInvoice('1', [{ description: 'A', quantity: 1, unitPrice: 100 }], 0.10);
    const updated = c.addItem('1', { description: 'B', quantity: 1, unitPrice: 50 });
    expect(updated!.subtotal).toBe(150);
    expect(updated!.total).toBeCloseTo(165.0);
  }
  // add item non-existent
  {
    const c = create();
    expect(c.addItem('nope', { description: 'X', quantity: 1, unitPrice: 1 })).toBeUndefined();
  }
  // apply discount
  {
    const c = create();
    c.createInvoice('1', [{ description: 'A', quantity: 3, unitPrice: 111.00 }], 0.10);
    const d = c.applyDiscount('1', 10);
    expect(d).toBeDefined();
    expect(d!.subtotal).toBeCloseTo(299.70);
    expect(d!.tax).toBeCloseTo(29.97);
    expect(d!.total).toBeCloseTo(329.67);
    expect(d!.discount).toBe(10);
  }
  // apply discount 0%
  {
    const c = create();
    c.createInvoice('1', [{ description: 'A', quantity: 1, unitPrice: 100 }], 0.10);
    const d = c.applyDiscount('1', 0);
    expect(d!.total).toBeCloseTo(110.0);
  }
  // apply discount non-existent
  {
    const c = create();
    expect(c.applyDiscount('nope', 10)).toBeUndefined();
  }
  // precision: 3 * 111.00 with 10% tax should round to 366.30
  {
    const c = create();
    const inv = c.createInvoice('p', [{ description: 'X', quantity: 3, unitPrice: 111.00 }], 0.10);
    expect(Math.round(inv.total * 100) / 100).toBe(366.30);
  }
  // precision: multiple items
  {
    const c = create();
    const inv = c.createInvoice('q', [
      { description: 'A', quantity: 1, unitPrice: 0.1 },
      { description: 'B', quantity: 1, unitPrice: 0.2 },
    ], 0.10);
    expect(inv.subtotal).toBeCloseTo(0.3);
    expect(inv.total).toBeCloseTo(0.33);
  }
}

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", { encoding: "utf-8", stdio: "pipe" });
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
