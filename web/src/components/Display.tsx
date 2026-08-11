interface Props {
  result: number | null;
  error: string | null;
  isLoading: boolean;
}

/**
 * Formats a float64 for display.
 *
 * The backend returns the raw IEEE 754 value, so 0.1 + 0.2 arrives as
 * 0.30000000000000004. Rounding is a presentation concern and is applied
 * here rather than in the API, which stays precise.
 */
function format(value: number): string {
  const rounded = Number(value.toPrecision(12));
  return Number.isInteger(rounded) ? String(rounded) : String(rounded);
}

export function Display({ result, error, isLoading }: Props) {
  return (
    <output className="display" aria-live="polite">
      {isLoading && <span className="display__hint">Calculating…</span>}
      {!isLoading && error !== null && <span className="display__error">{error}</span>}
      {!isLoading && error === null && result !== null && (
        <span className="display__result">{format(result)}</span>
      )}
      {!isLoading && error === null && result === null && (
        <span className="display__hint">Enter a calculation</span>
      )}
    </output>
  );
}