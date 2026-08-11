import type {
  ApiErrorCode,
  ApiErrorResponse,
  CalculateRequest,
  CalculateResponse,
} from './types';

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

/** Requests are aborted after this many milliseconds. */
const REQUEST_TIMEOUT_MS = 10_000;

/**
 * An error carrying the backend's machine-readable code, so the UI can
 * react to the failure kind rather than parsing message strings.
 *
 * Failures that never reach the backend (network down, timeout) are
 * reported with the synthetic code 'NETWORK_ERROR'.
 */
export class ApiError extends Error {
  readonly code: ApiErrorCode | 'NETWORK_ERROR';
  readonly status: number;

  constructor(code: ApiErrorCode | 'NETWORK_ERROR', message: string, status = 0) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

/**
 * Calls POST /api/v1/calculate.
 *
 * Resolves with the result, or rejects with an ApiError. Callers only need
 * to handle one error type regardless of where the failure originated.
 */
export async function calculate(
  request: CalculateRequest,
  signal?: AbortSignal,
): Promise<CalculateResponse> {
  const timeout = AbortSignal.timeout(REQUEST_TIMEOUT_MS);

  let response: Response;
  try {
    response = await fetch(`${BASE_URL}/api/v1/calculate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal: signal ? AbortSignal.any([signal, timeout]) : timeout,
    });
  } catch (cause) {
    // fetch only rejects when the request never completed: the server is
    // unreachable, the connection dropped, or the request was aborted.
    // HTTP error statuses resolve normally and are handled below.
    if (cause instanceof DOMException && cause.name === 'AbortError') {
      throw new ApiError('NETWORK_ERROR', 'The request timed out. Please try again.');
    }
    throw new ApiError('NETWORK_ERROR', 'Cannot reach the server. Is the backend running?');
  }

  if (!response.ok) {
    throw await toApiError(response);
  }

  return (await response.json()) as CalculateResponse;
}

/** Builds an ApiError from a non-2xx response. */
async function toApiError(response: Response): Promise<ApiError> {
  try {
    const body = (await response.json()) as ApiErrorResponse;
    if (body?.error?.code) {
      return new ApiError(body.error.code, body.error.message, response.status);
    }
  } catch {
    // The body was not the expected JSON envelope; fall through.
  }
  return new ApiError('INTERNAL_ERROR', `Request failed with status ${response.status}`, response.status);
}