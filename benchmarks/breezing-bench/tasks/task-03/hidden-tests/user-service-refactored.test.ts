import { describe, it, expect, beforeEach } from 'vitest';
import { UserService } from '../user-service';

describe('UserService - Refactored (hidden tests)', () => {
  let service: UserService;

  beforeEach(() => {
    service = new UserService();
    service.reset();
  });

  describe('Original tests (must still pass)', () => {
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

  describe('Refactoring validation', () => {
    it('should verify that auth module exists separately', async () => {
      // Check if auth module was extracted
      try {
        const authModule = await import('../auth');
        expect(authModule).toBeDefined();
      } catch {
        throw new Error('Auth module should be extracted to src/auth.ts');
      }
    });

    it('should verify that profile module exists separately', async () => {
      // Check if profile module was extracted
      try {
        const profileModule = await import('../profile');
        expect(profileModule).toBeDefined();
      } catch {
        throw new Error('Profile module should be extracted to src/profile.ts');
      }
    });

    it('should verify that notification module exists separately', async () => {
      // Check if notification module was extracted
      try {
        const notificationModule = await import('../notification');
        expect(notificationModule).toBeDefined();
      } catch {
        throw new Error('Notification module should be extracted to src/notification.ts');
      }
    });

    it('should handle session creation and verification', () => {
      const user = service.createUser('test@example.com', 'password123', 'Test User');
      const token = service.createSession(user.id);
      expect(token).toBeDefined();

      const verifiedUser = service.verifyToken(token);
      expect(verifiedUser?.id).toBe(user.id);
    });

    it('should send welcome email with notification', () => {
      const user = service.createUser('test@example.com', 'password123', 'Test User');
      service.sendWelcomeEmail(user.id);
      const notifications = service.getNotifications(user.id);
      expect(notifications.length).toBeGreaterThan(0);
      expect(notifications.some((n) => n.message.includes('Welcome'))).toBe(true);
    });

    it('should verify user email', () => {
      const user = service.createUser('test@example.com', 'password123', 'Test User');
      expect(user.verified).toBe(false);
      service.verifyUserEmail(user.id);
      const profile = service.getProfile(user.id);
      expect(profile?.verified).toBe(true);
    });

    it('should validate email format correctly', () => {
      expect(() => service.createUser('invalid-email', 'password123', 'Test')).toThrow('Invalid email');
    });
  });
});
