export interface Todo {
  id: string;
  title: string;
  description?: string;
  completed: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateTodoInput {
  title: string;
  description?: string;
}

export interface UpdateTodoInput {
  title?: string;
  description?: string;
  completed?: boolean;
}

export interface TodoStore {
  create(input: CreateTodoInput): Todo;
  getById(id: string): Todo | undefined;
  getAll(): Todo[];
  update(id: string, input: UpdateTodoInput): Todo | undefined;
  delete(id: string): boolean;
}
