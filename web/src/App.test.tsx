import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import App from './App';
import { ApiError } from './api/client';

vi.mock('./api/client', async () => {
  const actual = await vi.importActual<typeof import('./api/client')>('./api/client');
  return { ...actual, calculate: vi.fn() };
});

const { calculate } = await import('./api/client');
const mockCalculate = vi.mocked(calculate);

beforeEach(() => {
  mockCalculate.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('App', () => {
  it('shows a hint before any calculation', () => {
    render(<App />);

    expect(screen.getByText('Enter a calculation')).toBeInTheDocument();
  });

  it('calculates and displays the result', async () => {
    const user = userEvent.setup();
    mockCalculate.mockResolvedValue({ operation: 'add', a: 2, b: 3, result: 5 });

    render(<App />);

    await user.type(screen.getByLabelText('First number'), '2');
    await user.type(screen.getByLabelText('Second number'), '3');
    await user.click(screen.getByRole('button', { name: 'Calculate' }));

    expect(await screen.findByText('5')).toBeInTheDocument();
  });

  it('hides the second operand for unary operations', async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(screen.getByLabelText('Second number')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'sqrt' }));

    expect(screen.queryByLabelText('Second number')).not.toBeInTheDocument();
    expect(screen.getByLabelText('Number')).toBeInTheDocument();
  });

  it('displays a friendly message when the backend rejects the calculation', async () => {
    const user = userEvent.setup();
    mockCalculate.mockRejectedValue(
      new ApiError('DIVISION_BY_ZERO', 'division by zero is undefined', 422),
    );

    render(<App />);

    await user.click(screen.getByRole('button', { name: 'divide' }));
    await user.type(screen.getByLabelText('First number'), '10');
    await user.type(screen.getByLabelText('Second number'), '0');
    await user.click(screen.getByRole('button', { name: 'Calculate' }));

    expect(await screen.findByText('Cannot divide by zero.')).toBeInTheDocument();
  });

  it('validates before sending a request', async () => {
    const user = userEvent.setup();

    render(<App />);

    await user.click(screen.getByRole('button', { name: 'Calculate' }));

    expect(await screen.findByText('Enter the first number.')).toBeInTheDocument();
    expect(mockCalculate).not.toHaveBeenCalled();
  });

  it('marks the selected operation as pressed', async () => {
    const user = userEvent.setup();

    render(<App />);

    expect(screen.getByRole('button', { name: 'add' })).toHaveAttribute('aria-pressed', 'true');

    await user.click(screen.getByRole('button', { name: 'multiply' }));

    expect(screen.getByRole('button', { name: 'multiply' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'add' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('clears the form', async () => {
    const user = userEvent.setup();
    mockCalculate.mockResolvedValue({ operation: 'add', a: 2, b: 3, result: 5 });

    render(<App />);

    await user.type(screen.getByLabelText('First number'), '2');
    await user.type(screen.getByLabelText('Second number'), '3');
    await user.click(screen.getByRole('button', { name: 'Calculate' }));
    await screen.findByText('5');

    await user.click(screen.getByRole('button', { name: 'Clear' }));

    expect(screen.getByLabelText('First number')).toHaveValue('');
    expect(screen.getByText('Enter a calculation')).toBeInTheDocument();
  });

  it('rounds floating-point noise out of the display', async () => {
    const user = userEvent.setup();
    // 0.1 + 0.2 is 0.30000000000000004 in IEEE 754; 
    mockCalculate.mockResolvedValue({
      operation: 'add',
      a: 0.1,
      b: 0.2,
      result: 0.30000000000000004,
    });

    render(<App />);

    await user.type(screen.getByLabelText('First number'), '0.1');
    await user.type(screen.getByLabelText('Second number'), '0.2');
    await user.click(screen.getByRole('button', { name: 'Calculate' }));

    expect(await screen.findByText('0.3')).toBeInTheDocument();
  });
});