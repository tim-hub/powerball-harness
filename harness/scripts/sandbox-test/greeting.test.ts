/**
 * Sandbox Test - Greeting Utility Tests
 * Created by task-worker for /work --full testing
 */

import { describe, it, expect } from "vitest";
import { greet, greetWithTime } from "./greeting";

describe("greet", () => {
  it("should return greeting for a valid name", () => {
    expect(greet("World")).toBe("Hello, World!");
    expect(greet("Claude")).toBe("Hello, Claude!");
  });

  it("should trim whitespace from name", () => {
    expect(greet("  Alice  ")).toBe("Hello, Alice!");
    expect(greet("\tBob\n")).toBe("Hello, Bob!");
  });

  it("should return default greeting for empty string", () => {
    expect(greet("")).toBe("Hello, stranger!");
    expect(greet("   ")).toBe("Hello, stranger!");
  });

  it("should throw error for null", () => {
    expect(() => greet(null)).toThrow("Name cannot be null or undefined");
  });

  it("should throw error for undefined", () => {
    expect(() => greet(undefined)).toThrow("Name cannot be null or undefined");
  });
});

describe("greetWithTime", () => {
  it("should return morning greeting before noon", () => {
    expect(greetWithTime("Alice", 8)).toBe("Good morning, Alice!");
    expect(greetWithTime("Bob", 11)).toBe("Good morning, Bob!");
  });

  it("should return afternoon greeting between noon and 6pm", () => {
    expect(greetWithTime("Alice", 12)).toBe("Good afternoon, Alice!");
    expect(greetWithTime("Bob", 17)).toBe("Good afternoon, Bob!");
  });

  it("should return evening greeting after 6pm", () => {
    expect(greetWithTime("Alice", 18)).toBe("Good evening, Alice!");
    expect(greetWithTime("Bob", 23)).toBe("Good evening, Bob!");
  });

  it("should use default name for empty string", () => {
    expect(greetWithTime("", 10)).toBe("Good morning, stranger!");
    expect(greetWithTime("   ", 15)).toBe("Good afternoon, stranger!");
  });
});
