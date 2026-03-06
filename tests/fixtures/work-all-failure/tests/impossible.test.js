import test from "node:test";
import assert from "node:assert/strict";

import { add } from "../src/math.js";

test("baseline expectation", () => {
  assert.equal(add(2, 2), 4);
});

test("impossible expectation", () => {
  assert.equal(add(2, 2), 5);
});
