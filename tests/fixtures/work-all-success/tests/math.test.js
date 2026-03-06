import test from "node:test";
import assert from "node:assert/strict";

import { add, subtract } from "../src/math.js";
import { formatResult } from "../src/format.js";

test("add returns the arithmetic sum", () => {
  assert.equal(add(2, 3), 5);
});

test("subtract returns the arithmetic difference", () => {
  assert.equal(subtract(7, 2), 5);
});

test("formatResult renders a stable label", () => {
  assert.equal(formatResult("total", 5), "total: 5");
});
