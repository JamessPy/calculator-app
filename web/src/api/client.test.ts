import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError, calculate } from './client';

const mockFetch = vi.fn();

beforeEach(() => {
  vi.stubGlobal('fetch', mockFetch);
  mockFetch.mockReset();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

/** Builds a Response-like object with just the fields the client reads. */
function jsonResponse(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  } as Response;
}

describe('calculate', () => {
  it('posts JSON to the calculate endpoint', async () => {
    mockFetch.mockResolvedValue(
      jsonResponse(200, { operation: 'add', a: 2, b: 3, result: 5 }),
    );

    const result = await calculate({ operation: 'add', a: 2, b: 3 });

    expect(result.result).toBe(5);

    const [url, init] = mockFetch.mock.calls[0];
    expect(url).toContain('/api/v1/calculate');
    expect(init.method).toBe('POST');
    expect(init.headers['Content-Type']).toBe('application/json');
    expect(JSON.parse(init.body)).toEqual({ operation: 'add', a: 2, b: 3 });
  });

  it('omits the second operand for unary operations', async () => {
    mockFetch.mockResolvedValue(jsonResponse(200, { operation: 'sqrt', a: 144, result: 12 }));

    await calculate({ operation: 'sqrt', a: 144 });

    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toEqual({ operation: 'sqrt', a: 144 });
    expect('b' in body).toBe(false);
  });

  it('maps a backend error envelope onto ApiError', async () => {
    mockFetch.mockResolvedValue(
      jsonResponse(422, {
        error: { code: 'DIVISION_BY_ZERO', message: 'division by zero is undefined' },
      }),
    );

    await expect(calculate({ operation: 'divide', a: 1, b: 0 })).rejects.toMatchObject({
      code: 'DIVISION_BY_ZERO',
      status: 422,
    });
  });

  it('falls back to a generic error when the body is not the expected envelope', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 502,
      json: async () => {
        throw new SyntaxError('not JSON');
      },
    } as unknown as Response);

    await expect(calculate({ operation: 'add', a: 1, b: 2 })).rejects.toMatchObject({
      code: 'INTERNAL_ERROR',
      status: 502,
    });
  });

  it('reports an unreachable server as a network error', async () => {
    mockFetch.mockRejectedValue(new TypeError('Failed to fetch'));

    await expect(calculate({ operation: 'add', a: 1, b: 2 })).rejects.toMatchObject({
      code: 'NETWORK_ERROR',
    });
  });

  it('reports an aborted request as a timeout', async () => {
    mockFetch.mockRejectedValue(new DOMException('The operation was aborted.', 'AbortError'));

    await expect(calculate({ operation: 'add', a: 1, b: 2 })).rejects.toMatchObject({
      code: 'NETWORK_ERROR',
    });
    await expect(calculate({ operation: 'add', a: 1, b: 2 })).rejects.toThrow(/timed out/);
  });

  it('rejects with an ApiError instance so callers can narrow the type', async () => {
    mockFetch.mockRejectedValue(new TypeError('Failed to fetch'));

    await expect(calculate({ operation: 'add', a: 1, b: 2 })).rejects.toBeInstanceOf(ApiError);
  });
});