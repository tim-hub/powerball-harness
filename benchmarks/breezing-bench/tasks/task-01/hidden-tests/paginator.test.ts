import { describe, it, expect } from 'vitest';
import { paginate } from '../paginator';

describe('paginate - hidden tests', () => {
  const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12];

  it('should return first page with correct items', () => {
    const result = paginate(items, 1, 3);
    expect(result.items).toEqual([1, 2, 3]);
    expect(result.page).toBe(1);
    expect(result.totalPages).toBe(4);
    expect(result.totalItems).toBe(12);
  });

  it('should return middle page correctly', () => {
    const result = paginate(items, 2, 3);
    expect(result.items).toEqual([4, 5, 6]);
    expect(result.hasNext).toBe(true);
    expect(result.hasPrev).toBe(true);
  });

  it('should return last page correctly', () => {
    const result = paginate(items, 4, 3);
    expect(result.items).toEqual([10, 11, 12]);
    expect(result.hasNext).toBe(false);
    expect(result.hasPrev).toBe(true);
  });

  it('should handle empty array', () => {
    const result = paginate([], 1, 10);
    expect(result.items).toEqual([]);
    expect(result.totalPages).toBe(0);
    expect(result.hasNext).toBe(false);
    expect(result.hasPrev).toBe(false);
  });

  it('should handle page beyond total pages', () => {
    const result = paginate(items, 10, 3);
    expect(result.items).toEqual([]);
    expect(result.hasNext).toBe(false);
  });

  it('should handle itemsPerPage of 1', () => {
    const result = paginate([1, 2, 3], 2, 1);
    expect(result.items).toEqual([2]);
    expect(result.totalPages).toBe(3);
  });

  it('should verify page 1 has no previous page', () => {
    const result = paginate(items, 1, 5);
    expect(result.hasPrev).toBe(false);
  });

  it('should verify last page has no next page', () => {
    const result = paginate(items, 3, 5);
    expect(result.items).toEqual([11, 12]);
    expect(result.hasNext).toBe(false);
  });

  it('should calculate totalPages correctly for exact division', () => {
    const result = paginate([1, 2, 3, 4, 5, 6], 1, 3);
    expect(result.totalPages).toBe(2);
  });

  it('should handle single item per page boundary', () => {
    const result = paginate([1, 2, 3, 4, 5], 5, 1);
    expect(result.items).toEqual([5]);
    expect(result.hasNext).toBe(false);
    expect(result.hasPrev).toBe(true);
  });
});
