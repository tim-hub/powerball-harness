import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ApiClient } from '../api-client';

describe('ApiClient - Error Handling (hidden tests)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn();
  });

  describe('Retry on network errors', () => {
    it('should retry on network error', async () => {
      const mockFetch = vi.fn()
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => ({ success: true }),
        });

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        retries: 3,
      });

      const result = await client.get('/test');

      expect(mockFetch).toHaveBeenCalledTimes(3);
      expect(result.status).toBe(200);
      expect(result.data).toEqual({ success: true });
    });

    it('should throw after max retries exceeded', async () => {
      const mockFetch = vi.fn()
        .mockRejectedValue(new Error('Network error'));

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        retries: 2,
      });

      await expect(client.get('/test')).rejects.toThrow();
      expect(mockFetch).toHaveBeenCalledTimes(3); // 1 initial + 2 retries
    });
  });

  describe('Timeout handling', () => {
    it('should timeout after specified duration', async () => {
      const mockFetch = vi.fn().mockImplementation(() =>
        new Promise((resolve) => setTimeout(resolve, 2000))
      );

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        timeout: 100,
      });

      await expect(client.get('/test')).rejects.toThrow();
    });

    it('should use custom timeout value', async () => {
      const mockFetch = vi.fn().mockImplementation(() =>
        new Promise((resolve) => setTimeout(() => resolve({
          ok: true,
          status: 200,
          json: async () => ({ success: true }),
        }), 50))
      );

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        timeout: 200,
      });

      const result = await client.get('/test');
      expect(result.status).toBe(200);
    });
  });

  describe('HTTP status code handling', () => {
    it('should NOT retry on 4xx errors', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({ error: 'Not found' }),
      });

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        retries: 3,
      });

      const result = await client.get('/test');

      expect(mockFetch).toHaveBeenCalledTimes(1); // No retries for 4xx
      expect(result.status).toBe(404);
    });

    it('should retry on 5xx errors', async () => {
      const mockFetch = vi.fn()
        .mockResolvedValueOnce({
          ok: false,
          status: 500,
          json: async () => ({ error: 'Server error' }),
        })
        .mockResolvedValueOnce({
          ok: false,
          status: 503,
          json: async () => ({ error: 'Service unavailable' }),
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => ({ success: true }),
        });

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        retries: 3,
      });

      const result = await client.get('/test');

      expect(mockFetch).toHaveBeenCalledTimes(3);
      expect(result.status).toBe(200);
    });
  });

  describe('Success cases', () => {
    it('should return response on first success', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ message: 'Success' }),
      });

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        retries: 3,
      });

      const result = await client.get('/test');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      expect(result.data).toEqual({ message: 'Success' });
    });
  });

  describe('POST method', () => {
    it('should retry POST on network error', async () => {
      const mockFetch = vi.fn()
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          status: 201,
          json: async () => ({ id: 123 }),
        });

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        retries: 2,
      });

      const result = await client.post('/users', { name: 'Test' });

      expect(mockFetch).toHaveBeenCalledTimes(2);
      expect(result.status).toBe(201);
    });

    it('should pass correct body in POST', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ success: true }),
      });

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
      });

      const body = { name: 'Test User' };
      await client.post('/users', body);

      expect(mockFetch).toHaveBeenCalledWith(
        'https://api.example.com/users',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        })
      );
    });
  });

  describe('AbortController integration', () => {
    it('should support request cancellation', async () => {
      const mockFetch = vi.fn().mockImplementation(() =>
        new Promise((resolve) => setTimeout(resolve, 1000))
      );

      global.fetch = mockFetch;

      const client = new ApiClient({
        baseUrl: 'https://api.example.com',
        timeout: 50,
      });

      await expect(client.get('/test')).rejects.toThrow();

      // Verify AbortController was used
      expect(mockFetch).toHaveBeenCalled();
      const callArgs = mockFetch.mock.calls[0];
      if (callArgs[1]?.signal) {
        expect(callArgs[1].signal).toBeInstanceOf(AbortSignal);
      }
    });
  });
});
