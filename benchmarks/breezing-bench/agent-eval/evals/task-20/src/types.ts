export interface IDoublyLinkedList<T> {
  append(value: T): void;
  prepend(value: T): void;
  delete(value: T): boolean;
  find(predicate: (value: T) => boolean): T | undefined;
  toArray(): T[];
  size(): number;
  reverse(): void;
}
