import type { User } from './types';

const users = new Map<string, User>();

export const db = {
  getUser: (id: string): User | undefined => users.get(id),
  getUserByEmail: (email: string): User | undefined => {
    return Array.from(users.values()).find((u) => u.email === email);
  },
  saveUser: (user: User): void => {
    users.set(user.id, user);
  },
  clear: (): void => users.clear(),
};
