import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../api/client';
import { useCalculator } from './useCalculator';

// The API client is mocked so the hook can be tested without a backend.
vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return { ...actual, calculate: vi.fn() };
});

const { calculate } = await import('../api/client');
const mockCalculate = vi.mocked(calculate);

beforeEach(() => {
  mockCalculate.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('useCalculator', () => {
  it('starts empty', () => {
    const { result } = renderHook(() => useCalculator());

    expect(result.current.a).toBe('');
    expect(result.current.b).toBe('');
    expect(result.current.operation).toBe('add');
    expect(result.current.result).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('sends the operands and stores the result', async () => {
    mockCalculate.mockResolvedValue({ operation: 'add', a: 2, b: 3, result: 5 });

    const { result } = renderHook(() => useCalculator());

    act(() => {
      result.current.setA('2');
      result.current.setB('3');
    });
    await act(async () => {
      await result.current.submit();
    });

    expect(mockCalculate).toHaveBeenCalledWith({ operation: 'add', a: 2, b: 3 });
    expect(result.current.result).toBe(5);
    expect(result.current.error).toBeNull();
  });

  it('omits the second operand for unary operations', async () => {
    mockCalculate.mockResolvedValue({ operation: 'sqrt', a: 144, result: 12 });

    const { result } = renderHook(() => useCalculator());

    act(() => {
      result.current.changeOperation('sqrt');
      result.current.setA('144');
    });
    await act(async () => {
      await result.current.submit();
    });

    expect(mockCalculate).toHaveBeenCalledWith({ operation: 'sqrt', a: 144 });
    expect(result.current.needsSecondOperand).toBe(false);
  });

  describe('client-side validation', () => {
    it('rejects an empty first operand', async () => {
      const { result } = renderHook(() => useCalculator());

      await act(async () => {
        await result.current.submit();
      });

      expect(result.current.error).toBe('Enter the first number.');
      expect(mockCalculate).not.toHaveBeenCalled();
    });

    it('rejects a non-numeric operand', async () => {
      const { result } = renderHook(() => useCalculator());

      act(() => {
        result.current.setA('abc');
        result.current.setB('3');
      });
      await act(async () => {
        await result.current.submit();
      });

      expect(result.current.error).toContain('not a valid number');
      expect(mockCalculate).not.toHaveBeenCalled();
    });

    it('rejects a missing second operand for binary operations', async () => {
      const { result } = renderHook(() => useCalculator());

      act(() => {
        result.current.setA('2');
      });
      await act(async () => {
        await result.current.submit();
      });

      expect(result.current.error).toBe('Enter the second number.');
      expect(mockCalculate).not.toHaveBeenCalled();
    });

    it('accepts zero as an operand', async () => {
      mockCalculate.mockResolvedValue({ operation: 'add', a: 0, b: 0, result: 0 });

      const { result } = renderHook(() => useCalculator());

      act(() => {
        result.current.setA('0');
        result.current.setB('0');
      });
      await act(async () => {
        await result.current.submit();
      });

      expect(mockCalculate).toHaveBeenCalledWith({ operation: 'add', a: 0, b: 0 });
      expect(result.current.result).toBe(0);
    });
  });

  describe('backend errors', () => {
    it('translates the error code into a user-facing message', async () => {
      mockCalculate.mockRejectedValue(
        new ApiError('DIVISION_BY_ZERO', 'division by zero is undefined', 422),
      );

      const { result } = renderHook(() => useCalculator());

      act(() => {
        result.current.changeOperation('divide');
        result.current.setA('10');
        result.current.setB('0');
      });
      await act(async () => {
        await result.current.submit();
      });

      expect(result.current.error).toBe('Cannot divide by zero.');
      expect(result.current.result).toBeNull();
    });

    it('surfaces network failures', async () => {
      mockCalculate.mockRejectedValue(
        new ApiError('NETWORK_ERROR', 'Cannot reach the server. Is the backend running?'),
      );

      const { result } = renderHook(() => useCalculator());

      act(() => {
        result.current.setA('1');
        result.current.setB('2');
      });
      await act(async () => {
        await result.current.submit();
      });

      expect(result.current.error).toContain('Cannot reach the server');
    });

    it('clears the loading flag after a failure', async () => {
      mockCalculate.mockRejectedValue(new ApiError('INTERNAL_ERROR', 'boom', 500));

      const { result } = renderHook(() => useCalculator());

      act(() => {
        result.current.setA('1');
        result.current.setB('2');
      });
      await act(async () => {
        await result.current.submit();
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
    });
  });

  it('clears the form', async () => {
    mockCalculate.mockResolvedValue({ operation: 'add', a: 2, b: 3, result: 5 });

    const { result } = renderHook(() => useCalculator());

    act(() => {
      result.current.setA('2');
      result.current.setB('3');
    });
    await act(async () => {
      await result.current.submit();
    });
    act(() => {
      result.current.clear();
    });

    expect(result.current.a).toBe('');
    expect(result.current.b).toBe('');
    expect(result.current.result).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('discards a stale result when the operation changes', async () => {
    mockCalculate.mockResolvedValue({ operation: 'add', a: 2, b: 3, result: 5 });

    const { result } = renderHook(() => useCalculator());

    act(() => {
      result.current.setA('2');
      result.current.setB('3');
    });
    await act(async () => {
      await result.current.submit();
    });
    expect(result.current.result).toBe(5);

    act(() => {
      result.current.changeOperation('multiply');
    });

    expect(result.current.result).toBeNull();
  });
  it.each([
      ['RESULT_NOT_FINITE', 'The result is too large to display.'],
      ['OPERAND_REQUIRED', 'This operation needs a second number.'],
      ['UNSUPPORTED_OPERATION', 'That operation is not supported.'],
      ['NEGATIVE_SQUARE_ROOT', 'Cannot take the square root of a negative number.'],
    ] as const)('maps %s to a user-facing message', async (code, expected) => {
      mockCalculate.mockRejectedValue(new ApiError(code, 'backend wording', 422));

      const { result } = renderHook(() => useCalculator());

      act(() => {
        result.current.setA('1');
        result.current.setB('2');
      });
      await act(async () => {
        await result.current.submit();
      });

      expect(result.current.error).toBe(expected);
    });
});