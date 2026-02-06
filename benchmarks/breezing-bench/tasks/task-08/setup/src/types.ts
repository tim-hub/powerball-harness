export interface RateLimiterOptions {
  maxTokens: number;       // バケット容量
  refillRate: number;      // 1秒あたりのリフィル数
  refillInterval?: number; // リフィル間隔 (ms), default 1000
}

export interface RateLimiter {
  tryConsume(tokens?: number): boolean;
  getAvailableTokens(): number;
  reset(): void;
}
