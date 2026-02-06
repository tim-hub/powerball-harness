import { describe, it, expect, beforeEach } from 'vitest';
import { UserService } from './user-service';

describe('UserService', () => {
  let service: UserService;

  beforeEach(() => {
    service = new UserService();
  });

  it('should create a new user', () => {
    const user = service.createUser('test@example.com', 'password123', 'Test User');
    expect(user.email).toBe('test@example.com');
    expect(user.name).toBe('Test User');
    expect(user.verified).toBe(false);
  });

  it('should verify password', () => {
    const user = service.createUser('test@example.com', 'password123', 'Test User');
    expect(service.verifyPassword(user, 'password123')).toBe(true);
    expect(service.verifyPassword(user, 'wrongpassword')).toBe(false);
  });

  it('should update profile', () => {
    const user = service.createUser('test@example.com', 'password123', 'Test User');
    const updated = service.updateProfile(user.id, { name: 'Updated Name' });
    expect(updated?.name).toBe('Updated Name');
  });

  it('should send notification', () => {
    const user = service.createUser('test@example.com', 'password123', 'Test User');
    service.sendNotification(user.id, 'Test notification');
    const notifications = service.getNotifications(user.id);
    expect(notifications).toHaveLength(1);
    expect(notifications[0].message).toBe('Test notification');
  });

  it('should log actions', () => {
    service.createUser('test@example.com', 'password123', 'Test User');
    const logs = service.getLogs();
    expect(logs.length).toBeGreaterThan(0);
    expect(logs.some((log) => log.includes('Creating user'))).toBe(true);
  });
});
