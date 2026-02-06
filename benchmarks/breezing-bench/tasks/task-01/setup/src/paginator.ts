export interface PaginationResult<T> {
  items: T[];
  page: number;
  totalPages: number;
  totalItems: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export function paginate<T>(items: T[], page: number, itemsPerPage: number): PaginationResult<T> {
  const totalItems = items.length;
  const totalPages = Math.ceil(totalItems / itemsPerPage);
  // BUG: off-by-one - should be (page - 1) * itemsPerPage
  const start = page * itemsPerPage;
  const end = start + itemsPerPage;

  return {
    items: items.slice(start, end),
    page,
    totalPages,
    totalItems,
    // BUG: should be page < totalPages
    hasNext: page <= totalPages,
    // BUG: should be page > 1
    hasPrev: page > 0,
  };
}
