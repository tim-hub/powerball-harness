import crypto from 'crypto';

interface User {
  id: string;
  email: string;
  passwordHash: string;
  name: string;
  verified: boolean;
}

interface Session {
  token: string;
  userId: string;
  expiresAt: Date;
}

interface Notification {
  userId: string;
  message: string;
  sentAt: Date;
}

export class UserService {
  private users = new Map<string, User>();
  private sessions = new Map<string, Session>();
  private notifications: Notification[] = [];
  private logs: string[] = [];

  // Authentication methods
  findUser(email: string): User | undefined {
    this.logAction(`Finding user: ${email}`);
    return Array.from(this.users.values()).find((u) => u.email === email);
  }

  createUser(email: string, password: string, name: string): User {
    this.logAction(`Creating user: ${email}`);
    if (!this.validateEmail(email)) {
      throw new Error('Invalid email');
    }
    if (this.findUser(email)) {
      throw new Error('User already exists');
    }
    const user: User = {
      id: crypto.randomUUID(),
      email,
      passwordHash: this.hashPassword(password),
      name,
      verified: false,
    };
    this.users.set(user.id, user);
    return user;
  }

  hashPassword(password: string): string {
    return crypto.createHash('sha256').update(password).digest('hex');
  }

  verifyPassword(user: User, password: string): boolean {
    this.logAction(`Verifying password for: ${user.email}`);
    return user.passwordHash === this.hashPassword(password);
  }

  createSession(userId: string): string {
    this.logAction(`Creating session for user: ${userId}`);
    const token = crypto.randomUUID();
    const session: Session = {
      token,
      userId,
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 24 hours
    };
    this.sessions.set(token, session);
    return token;
  }

  verifyToken(token: string): User | undefined {
    this.logAction(`Verifying token: ${token}`);
    const session = this.sessions.get(token);
    if (!session || session.expiresAt < new Date()) {
      return undefined;
    }
    return this.users.get(session.userId);
  }

  // Profile methods
  updateProfile(userId: string, updates: { name?: string; email?: string }): User | undefined {
    this.logAction(`Updating profile for user: ${userId}`);
    const user = this.users.get(userId);
    if (!user) return undefined;

    if (updates.email && !this.validateEmail(updates.email)) {
      throw new Error('Invalid email');
    }

    if (updates.name) user.name = updates.name;
    if (updates.email) user.email = updates.email;

    return user;
  }

  getProfile(userId: string): Omit<User, 'passwordHash'> | undefined {
    this.logAction(`Getting profile for user: ${userId}`);
    const user = this.users.get(userId);
    if (!user) return undefined;

    const { passwordHash, ...profile } = user;
    return profile;
  }

  verifyUserEmail(userId: string): boolean {
    this.logAction(`Verifying email for user: ${userId}`);
    const user = this.users.get(userId);
    if (!user) return false;

    user.verified = true;
    return true;
  }

  // Notification methods
  sendNotification(userId: string, message: string): void {
    this.logAction(`Sending notification to user: ${userId}`);
    const user = this.users.get(userId);
    if (!user) {
      throw new Error('User not found');
    }

    this.notifications.push({
      userId,
      message,
      sentAt: new Date(),
    });
  }

  getNotifications(userId: string): Notification[] {
    this.logAction(`Getting notifications for user: ${userId}`);
    return this.notifications.filter((n) => n.userId === userId);
  }

  sendWelcomeEmail(userId: string): void {
    this.logAction(`Sending welcome email to user: ${userId}`);
    const user = this.users.get(userId);
    if (!user) {
      throw new Error('User not found');
    }
    this.sendNotification(userId, `Welcome ${user.name}! Please verify your email.`);
  }

  sendPasswordResetEmail(email: string): void {
    this.logAction(`Sending password reset email to: ${email}`);
    const user = this.findUser(email);
    if (!user) {
      throw new Error('User not found');
    }
    this.sendNotification(user.id, 'Password reset link has been sent to your email.');
  }

  // Utility methods
  validateEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  logAction(action: string): void {
    this.logs.push(`[${new Date().toISOString()}] ${action}`);
  }

  getLogs(): string[] {
    return [...this.logs];
  }

  clearLogs(): void {
    this.logs = [];
  }

  // Testing utility
  reset(): void {
    this.users.clear();
    this.sessions.clear();
    this.notifications = [];
    this.logs = [];
  }
}
