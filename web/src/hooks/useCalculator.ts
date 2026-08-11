import { useCallback, useState } from 'react';
import { ApiError, calculate } from '../api/client';
import { isUnary, type Operation } from '../api/types';

/** Maps backend error codes to messages suitable for an end user. */
function userMessage(error: ApiError): string {
  switch (error.code) {
    case 'DIVISION_BY_ZERO':
      return 'Cannot divide by zero.';
    case 'NEGATIVE_SQUARE_ROOT':
      return 'Cannot take the square root of a negative number.';
    case 'RESULT_NOT_FINITE':
      return 'The result is too large to display.';
    case 'OPERAND_REQUIRED':
      return 'This operation needs a second number.';
    case 'UNSUPPORTED_OPERATION':
      return 'That operation is not supported.';
    case 'NETWORK_ERROR':
      return error.message;
    default:
      return 'Something went wrong. Please try again.';
  }
}

export interface CalculatorState {
  a: string;
  b: string;
  operation: Operation;
  result: number | null;
  error: string | null;
  isLoading: boolean;
}

export function useCalculator() {
  const [a, setA] = useState('');
  const [b, setB] = useState('');
  const [operation, setOperation] = useState<Operation>('add');
  const [result, setResult] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const needsSecondOperand = !isUnary(operation);

  /**
   * Client-side validation, kept deliberately minimal: it catches what the
   * user can see (empty or non-numeric input) and leaves the mathematical
   * rules to the backend, which is the single source of truth.
   */
  function validate(): string | null {
    if (a.trim() === '') return 'Enter the first number.';
    if (!Number.isFinite(Number(a))) return 'The first value is not a valid number.';

    if (needsSecondOperand) {
      if (b.trim() === '') return 'Enter the second number.';
      if (!Number.isFinite(Number(b))) return 'The second value is not a valid number.';
    }
    return null;
  }

  const submit = useCallback(async () => {
    const validationError = validate();
    if (validationError !== null) {
      setError(validationError);
      setResult(null);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await calculate({
        operation,
        a: Number(a),
        ...(needsSecondOperand ? { b: Number(b) } : {}),
      });
      setResult(response.result);
    } catch (err) {
      setResult(null);
      setError(err instanceof ApiError ? userMessage(err) : 'Something went wrong.');
    } finally {
      setIsLoading(false);
    }
  }, [a, b, operation, needsSecondOperand]);

  const clear = useCallback(() => {
    setA('');
    setB('');
    setResult(null);
    setError(null);
  }, []);

  function changeOperation(next: Operation) {
    setOperation(next);
    setResult(null);
    setError(null);
  }

  return {
    a,
    b,
    operation,
    result,
    error,
    isLoading,
    needsSecondOperand,
    setA,
    setB,
    changeOperation,
    submit,
    clear,
  };
}