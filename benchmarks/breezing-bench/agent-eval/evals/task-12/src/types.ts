export interface PriorityItem<T> {
  value: T;
  priority: number;
}

export interface IPriorityQueue<T> {
  enqueue(value: T, priority: number): void;
  dequeue(): T | undefined;
  peek(): T | undefined;
  size(): number;
  isEmpty(): boolean;
}
