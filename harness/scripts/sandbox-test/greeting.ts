/**
 * Sandbox Test - Greeting Utility
 * Created by task-worker for /work --full testing
 */

/**
 * Generates a greeting message for the given name
 * @param name - The name to greet
 * @returns A greeting string in the format "Hello, {name}!"
 * @throws Error if name is null or undefined
 */
export function greet(name: string | null | undefined): string {
  if (name === null || name === undefined) {
    throw new Error("Name cannot be null or undefined");
  }

  const trimmedName = name.trim();

  if (trimmedName === "") {
    return "Hello, stranger!";
  }

  return `Hello, ${trimmedName}!`;
}

/**
 * Generates a personalized greeting with time of day
 * @param name - The name to greet
 * @param hour - The hour of the day (0-23)
 * @returns A time-appropriate greeting
 */
export function greetWithTime(name: string, hour: number): string {
  const timeGreeting =
    hour < 12 ? "Good morning" : hour < 18 ? "Good afternoon" : "Good evening";

  const trimmedName = name.trim() || "stranger";
  return `${timeGreeting}, ${trimmedName}!`;
}
