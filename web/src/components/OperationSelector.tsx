import { OPERATIONS, type Operation } from '../api/types';

/** Symbols shown to the user; the API contract uses the string keys. */
const LABELS: Record<Operation, string> = {
  add: '+',
  subtract: '−',
  multiply: '×',
  divide: '÷',
  power: 'xʸ',
  sqrt: '√',
  percentage: '%',
};

interface Props {
  value: Operation;
  onChange: (operation: Operation) => void;
  disabled?: boolean;
}

export function OperationSelector({ value, onChange, disabled }: Props) {
  return (
    <div className="operations" role="group" aria-label="Operation">
      {OPERATIONS.map((op) => (
        <button
          key={op}
          type="button"
          className={op === value ? 'op op--active' : 'op'}
          onClick={() => onChange(op)}
          disabled={disabled}
          aria-pressed={op === value}
          aria-label={op}
        >
          {LABELS[op]}
        </button>
      ))}
    </div>
  );
}