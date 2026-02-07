export interface IStack<T> {
  push(item: T): void;
  pop(): T | undefined;
  peek(): T | undefined;
  size(): number;
  isEmpty(): boolean;
  toArray(): T[];
  clear(): void;
}
