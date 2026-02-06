import { test, expect } from "vitest";
import { execSync } from "child_process";
import { writeFileSync, mkdirSync, existsSync } from "fs";
import { join } from "path";

// Hidden test content - agent never sees this
const HIDDEN_TEST = `
import { describe, it, expect, beforeEach, vi } from 'vitest';

describe('RateLimiter', () => {
  // NOTE: \u30A8\u30FC\u30B8\u30A7\u30F3\u30C8\u304C\u5B9F\u88C5\u3059\u3079\u304D\u30D5\u30A1\u30A4\u30EB\u306E\u30D1\u30B9
  const IMPL_PATH = '../rate-limiter';

  let createRateLimiter: any;

  beforeEach(async () => {
    // \u5B9F\u88C5\u3092\u52D5\u7684\u306B\u30A4\u30F3\u30DD\u30FC\u30C8
    try {
      const module = await import(IMPL_PATH);
      createRateLimiter = module.createRateLimiter || module.default;
    } catch (e) {
      throw new Error(\\\`RateLimiter implementation not found at \\\${IMPL_PATH}. Please create the file.\\\`);
    }
  });

  it('should consume tokens up to maxTokens on initialization', () => {
    const limiter = createRateLimiter({ maxTokens: 10, refillRate: 5 });

    expect(limiter.tryConsume(5)).toBe(true);
    expect(limiter.getAvailableTokens()).toBe(5);
    expect(limiter.tryConsume(5)).toBe(true);
    expect(limiter.getAvailableTokens()).toBe(0);
  });

  it('should reject consumption when exceeding available tokens', () => {
    const limiter = createRateLimiter({ maxTokens: 5, refillRate: 1 });

    expect(limiter.tryConsume(6)).toBe(false);
    expect(limiter.getAvailableTokens()).toBe(5);
  });

  it('should refill tokens over time', async () => {
    vi.useFakeTimers();
    const limiter = createRateLimiter({ maxTokens: 10, refillRate: 5 });

    limiter.tryConsume(10);
    expect(limiter.getAvailableTokens()).toBe(0);

    await vi.advanceTimersByTimeAsync(1000);
    expect(limiter.getAvailableTokens()).toBe(5);

    await vi.advanceTimersByTimeAsync(1000);
    expect(limiter.getAvailableTokens()).toBe(10);

    vi.useRealTimers();
  });

  it('should handle burst consumption', () => {
    const limiter = createRateLimiter({ maxTokens: 100, refillRate: 10 });

    expect(limiter.tryConsume(100)).toBe(true);
    expect(limiter.getAvailableTokens()).toBe(0);
    expect(limiter.tryConsume(1)).toBe(false);
  });

  it('should reset to full capacity', () => {
    const limiter = createRateLimiter({ maxTokens: 10, refillRate: 5 });

    limiter.tryConsume(10);
    expect(limiter.getAvailableTokens()).toBe(0);

    limiter.reset();
    expect(limiter.getAvailableTokens()).toBe(10);
  });

  it('should allow consumption of 0 tokens', () => {
    const limiter = createRateLimiter({ maxTokens: 10, refillRate: 5 });

    expect(limiter.tryConsume(0)).toBe(true);
    expect(limiter.getAvailableTokens()).toBe(10);
  });

  it('should default to consuming 1 token when no argument', () => {
    const limiter = createRateLimiter({ maxTokens: 10, refillRate: 5 });

    expect(limiter.tryConsume()).toBe(true);
    expect(limiter.getAvailableTokens()).toBe(9);
  });

  it('should throw or reject on negative tokens', () => {
    const limiter = createRateLimiter({ maxTokens: 10, refillRate: 5 });

    expect(() => limiter.tryConsume(-1)).toThrow();
  });

  it('should handle concurrent access simulation', async () => {
    const limiter = createRateLimiter({ maxTokens: 5, refillRate: 1 });

    const results = await Promise.all([
      limiter.tryConsume(1),
      limiter.tryConsume(1),
      limiter.tryConsume(1),
      limiter.tryConsume(1),
      limiter.tryConsume(1),
      limiter.tryConsume(1),
    ]);

    const successCount = results.filter(r => r).length;
    expect(successCount).toBe(5);
  });

  it('should not refill when refillRate is 0', async () => {
    vi.useFakeTimers();
    const limiter = createRateLimiter({ maxTokens: 10, refillRate: 0 });

    limiter.tryConsume(5);
    expect(limiter.getAvailableTokens()).toBe(5);

    await vi.advanceTimersByTimeAsync(5000);
    expect(limiter.getAvailableTokens()).toBe(5);

    vi.useRealTimers();
  });

  it('should support custom refillInterval', async () => {
    vi.useFakeTimers();
    const limiter = createRateLimiter({
      maxTokens: 10,
      refillRate: 2,
      refillInterval: 500
    });

    limiter.tryConsume(10);
    expect(limiter.getAvailableTokens()).toBe(0);

    await vi.advanceTimersByTimeAsync(500);
    expect(limiter.getAvailableTokens()).toBe(2);

    await vi.advanceTimersByTimeAsync(500);
    expect(limiter.getAvailableTokens()).toBe(4);

    vi.useRealTimers();
  });
});
`;

test("hidden tests pass", () => {
  // Write hidden test file
  const hiddenDir = join(process.cwd(), "src", "__hidden_tests__");
  if (!existsSync(hiddenDir)) {
    mkdirSync(hiddenDir, { recursive: true });
  }
  writeFileSync(join(hiddenDir, "rate-limiter.test.ts"), HIDDEN_TEST);

  // Install dependencies if needed
  if (!existsSync(join(process.cwd(), "node_modules"))) {
    execSync("npm install", { stdio: "pipe" });
  }

  // Run hidden tests
  const result = execSync(
    "npx vitest run --reporter=json src/__hidden_tests__/",
    { encoding: "utf-8", stdio: "pipe" }
  );

  const report = JSON.parse(result);
  expect(report.numPassedTests).toBe(report.numTotalTests);
  expect(report.numTotalTests).toBeGreaterThanOrEqual(5);
});

test("typecheck passes", () => {
  const result = execSync("npx tsc --noEmit 2>&1 || true", {
    encoding: "utf-8",
    stdio: "pipe",
  });
  const tsErrors = (result).match(/error TS/g) || [];
  expect(tsErrors.length).toBe(0);
});
