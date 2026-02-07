export interface ITTLCache<V> {
  set(key: string, value: V, ttl?: number): void;
  get(key: string): V | undefined;
  delete(key: string): boolean;
  has(key: string): boolean;
  clear(): void;
  size(): number;
  getOrSet(key: string, factory: () => V | Promise<V>, ttl?: number): Promise<V>;
}
