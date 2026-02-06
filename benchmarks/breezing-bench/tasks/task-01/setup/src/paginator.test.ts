import { describe, it, expect } from 'vitest';
import { paginate } from './paginator';

describe('paginate', () => {
  const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

  it('should return first page correctly', () => {
    const result = paginate(items, 1, 3);
    expect(result.items).toEqual([1, 2, 3]);
    expect(result.page).toBe(1);
  });

  it('should return last page correctly', () => {
    const result = paginate(items, 4, 3);
    expect(result.items).toEqual([10]);
    expect(result.hasNext).toBe(false);
  });
});
