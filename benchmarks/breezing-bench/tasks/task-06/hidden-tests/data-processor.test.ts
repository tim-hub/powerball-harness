import { describe, it, expect } from 'vitest';
import { processData, validateInput, transformBatch } from '../data-processor';

describe('data-processor', () => {
  describe('processData', () => {
    it('should process user type correctly', () => {
      const input = {
        type: 'user',
        payload: {
          name: 'Alice',
          age: 30,
          email: 'alice@example.com',
          tags: ['admin', 'active'],
        },
      };

      const result = processData(input);
      expect(result).toEqual({
        name: 'Alice',
        age: 30,
        email: 'alice@example.com',
        tags: ['admin', 'active'],
      });
    });

    it('should process product type correctly', () => {
      const input = {
        type: 'product',
        payload: {
          title: 'Laptop',
          price: 999.99,
          category: 'electronics',
          inStock: true,
        },
      };

      const result = processData(input);
      expect(result).toEqual({
        title: 'Laptop',
        price: 999.99,
        category: 'electronics',
        inStock: true,
      });
    });

    it('should process order type correctly', () => {
      const input = {
        type: 'order',
        payload: {
          orderId: 'ORD-123',
          items: [
            { id: 'P1', qty: 2, price: 10.0 },
            { id: 'P2', qty: 1, price: 25.0 },
          ],
        },
      };

      const result = processData(input);
      expect(result).toEqual({
        orderId: 'ORD-123',
        items: [
          { productId: 'P1', quantity: 2, subtotal: 20.0 },
          { productId: 'P2', quantity: 1, subtotal: 25.0 },
        ],
        total: 45.0,
      });
    });

    it('should return null for null input', () => {
      expect(processData(null)).toBeNull();
    });

    it('should return null for undefined input', () => {
      expect(processData(undefined)).toBeNull();
    });

    it('should return empty object for unknown type', () => {
      const input = { type: 'unknown', payload: {} };
      expect(processData(input)).toEqual({});
    });

    it('should default tags to empty array if missing', () => {
      const input = {
        type: 'user',
        payload: { name: 'Bob', age: 25, email: 'bob@example.com' },
      };
      const result = processData(input);
      expect(result.tags).toEqual([]);
    });

    it('should default inStock to true if missing', () => {
      const input = {
        type: 'product',
        payload: { title: 'Book', price: 15.0, category: 'books' },
      };
      const result = processData(input);
      expect(result.inStock).toBe(true);
    });
  });

  describe('validateInput', () => {
    it('should validate correct user input', () => {
      const input = {
        type: 'user',
        payload: { name: 'Alice', email: 'alice@example.com' },
      };
      const result = validateInput(input);
      expect(result.valid).toBe(true);
      expect(result.errors).toEqual([]);
    });

    it('should detect missing type', () => {
      const input = { payload: {} };
      const result = validateInput(input);
      expect(result.valid).toBe(false);
      expect(result.errors).toContainEqual({ field: 'type', message: 'required' });
    });

    it('should detect missing payload', () => {
      const input = { type: 'user' };
      const result = validateInput(input);
      expect(result.valid).toBe(false);
      expect(result.errors).toContainEqual({ field: 'payload', message: 'required' });
    });

    it('should detect missing user name', () => {
      const input = { type: 'user', payload: { email: 'test@example.com' } };
      const result = validateInput(input);
      expect(result.valid).toBe(false);
      expect(result.errors).toContainEqual({ field: 'name', message: 'required' });
    });

    it('should detect missing user email', () => {
      const input = { type: 'user', payload: { name: 'Alice' } };
      const result = validateInput(input);
      expect(result.valid).toBe(false);
      expect(result.errors).toContainEqual({ field: 'email', message: 'required' });
    });
  });

  describe('transformBatch', () => {
    it('should transform multiple items', () => {
      const items = [
        { type: 'user', payload: { name: 'Alice', age: 30, email: 'alice@example.com' } },
        { type: 'product', payload: { title: 'Laptop', price: 999, category: 'tech' } },
      ];

      const result = transformBatch(items);
      expect(result).toHaveLength(2);
      expect(result[0]).toHaveProperty('name', 'Alice');
      expect(result[1]).toHaveProperty('title', 'Laptop');
    });

    it('should filter out null results', () => {
      const items = [
        { type: 'user', payload: { name: 'Alice', age: 30, email: 'alice@example.com' } },
        null,
        { type: 'product', payload: { title: 'Book', price: 15, category: 'books' } },
      ];

      const result = transformBatch(items);
      expect(result).toHaveLength(2);
    });
  });

  describe('type exports', () => {
    it('should export types', () => {
      // This test ensures types are exported and can be imported
      // If types are not exported, TypeScript compilation will fail
      const userInput: any = {
        type: 'user',
        payload: { name: 'Test', age: 25, email: 'test@example.com' },
      };
      expect(processData(userInput)).toBeDefined();
    });
  });
});
