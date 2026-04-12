export interface RateLimiterOptions {
  maxTokens: number;       // Bucket capacity
  refillRate: number;      // Refill count per second
  refillInterval?: number; // Refill interval (ms), default 1000
}

export interface RateLimiter {
  tryConsume(tokens?: number): boolean;
  getAvailableTokens(): number;
  reset(): void;
}
