import { describe, it, expect, beforeEach } from 'vitest';
import type { TodoStore } from '../types';

// This test will import the user's implementation
// For now, we assume it will be in todo-store.ts
describe('TodoStore - hidden tests', () => {
  let store: TodoStore;

  beforeEach(async () => {
    // Dynamic import - implementation should be in src/todo-store.ts
    const module = await import('../todo-store');
    const StoreClass = module.default || module.TodoStoreImpl || module.InMemoryTodoStore;
    store = new StoreClass();
  });

  describe('create', () => {
    it('should create a todo with all fields', () => {
      const todo = store.create({ title: 'Test Todo', description: 'Test Description' });
      expect(todo.id).toBeDefined();
      expect(todo.title).toBe('Test Todo');
      expect(todo.description).toBe('Test Description');
      expect(todo.completed).toBe(false);
      expect(todo.createdAt).toBeInstanceOf(Date);
      expect(todo.updatedAt).toBeInstanceOf(Date);
    });

    it('should create a todo without description', () => {
      const todo = store.create({ title: 'Test Todo' });
      expect(todo.description).toBeUndefined();
    });

    it('should generate unique IDs', () => {
      const todo1 = store.create({ title: 'Todo 1' });
      const todo2 = store.create({ title: 'Todo 2' });
      expect(todo1.id).not.toBe(todo2.id);
    });

    it('should reject empty title', () => {
      expect(() => store.create({ title: '' })).toThrow();
    });
  });

  describe('getById', () => {
    it('should return existing todo', () => {
      const created = store.create({ title: 'Test' });
      const found = store.getById(created.id);
      expect(found).toEqual(created);
    });

    it('should return undefined for non-existent id', () => {
      const found = store.getById('non-existent-id');
      expect(found).toBeUndefined();
    });
  });

  describe('getAll', () => {
    it('should return empty array initially', () => {
      const todos = store.getAll();
      expect(todos).toEqual([]);
    });

    it('should return all created todos', () => {
      store.create({ title: 'Todo 1' });
      store.create({ title: 'Todo 2' });
      store.create({ title: 'Todo 3' });
      const todos = store.getAll();
      expect(todos).toHaveLength(3);
    });
  });

  describe('update', () => {
    it('should update title', () => {
      const todo = store.create({ title: 'Original' });
      const updated = store.update(todo.id, { title: 'Updated' });
      expect(updated?.title).toBe('Updated');
    });

    it('should update completed status', () => {
      const todo = store.create({ title: 'Test' });
      const updated = store.update(todo.id, { completed: true });
      expect(updated?.completed).toBe(true);
    });

    it('should update updatedAt timestamp', () => {
      const todo = store.create({ title: 'Test' });
      const originalUpdatedAt = todo.updatedAt;

      // Wait a bit to ensure different timestamp
      setTimeout(() => {
        const updated = store.update(todo.id, { title: 'Updated' });
        expect(updated?.updatedAt.getTime()).toBeGreaterThan(originalUpdatedAt.getTime());
      }, 10);
    });

    it('should not modify createdAt', () => {
      const todo = store.create({ title: 'Test' });
      const updated = store.update(todo.id, { title: 'Updated' });
      expect(updated?.createdAt).toEqual(todo.createdAt);
    });

    it('should return undefined for non-existent id', () => {
      const updated = store.update('non-existent', { title: 'Updated' });
      expect(updated).toBeUndefined();
    });
  });

  describe('delete', () => {
    it('should delete existing todo', () => {
      const todo = store.create({ title: 'Test' });
      const deleted = store.delete(todo.id);
      expect(deleted).toBe(true);
      expect(store.getById(todo.id)).toBeUndefined();
    });

    it('should return false for non-existent id', () => {
      const deleted = store.delete('non-existent');
      expect(deleted).toBe(false);
    });

    it('should not affect other todos', () => {
      const todo1 = store.create({ title: 'Todo 1' });
      const todo2 = store.create({ title: 'Todo 2' });
      store.delete(todo1.id);
      expect(store.getById(todo2.id)).toBeDefined();
    });
  });
});
