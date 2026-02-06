export interface Book {
  id: string;
  title: string;
  author: string;
  isbn: string;
  publishedYear: number;
  genre: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateBookInput {
  title: string;
  author: string;
  isbn: string;
  publishedYear: number;
  genre: string;
}

export interface UpdateBookInput {
  title?: string;
  author?: string;
  publishedYear?: number;
  genre?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// NOTE: Express は使わず、純粋な関数ベースの API ハンドラを実装すること
// (テスト容易性のため)
export interface BookStore {
  create(input: CreateBookInput): Book;
  getById(id: string): Book | undefined;
  getAll(page?: number, pageSize?: number): PaginatedResponse<Book>;
  update(id: string, input: UpdateBookInput): Book | undefined;
  delete(id: string): boolean;
  findByIsbn(isbn: string): Book | undefined;
}
