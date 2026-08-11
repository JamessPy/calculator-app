/**
 * Contract shared with the Go backend.
 *
 * These types mirror internal/httpapi/dto.go. There is no code generation,
 * so they must be kept in sync manually; the backend is the source of truth.
 */

/** Operations supported by POST /api/v1/calculate. */
export const OPERATIONS = [
  'add',
  'subtract',
  'multiply',
  'divide',
  'power',
  'sqrt',
  'percentage',
] as const;

/** Union of the operation string literals: 'add' | 'subtract' | ... */
export type Operation = (typeof OPERATIONS)[number];

/** Operations that take a single operand; `b` is omitted for these. */
export const UNARY_OPERATIONS: readonly Operation[] = ['sqrt'];

export function isUnary(op: Operation): boolean {
  return UNARY_OPERATIONS.includes(op);
}

/** Request body of POST /api/v1/calculate. */
export interface CalculateRequest {
  operation: Operation;
  a: number;
  /** Omitted for unary operations. */
  b?: number;
}

/** Response body of a successful calculation. */
export interface CalculateResponse {
  operation: Operation;
  a: number;
  /** Absent for unary operations. */
  b?: number;
  result: number;
}

/**
 * Stable, machine-readable error codes returned by the backend.
 * The message may change; the code is part of the contract.
 */
export type ApiErrorCode =
  | 'INVALID_JSON'
  | 'VALIDATION_FAILED'
  | 'UNSUPPORTED_OPERATION'
  | 'DIVISION_BY_ZERO'
  | 'NEGATIVE_SQUARE_ROOT'
  | 'OPERAND_REQUIRED'
  | 'OPERAND_NOT_FINITE'
  | 'RESULT_NOT_FINITE'
  | 'INTERNAL_ERROR';

/** Error envelope returned by the backend for every failure. */
export interface ApiErrorResponse {
  error: {
    code: ApiErrorCode;
    message: string;
  };
}