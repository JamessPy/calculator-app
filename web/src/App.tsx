import { Display } from './components/Display';
import { OperandInput } from './components/OperandInput';
import { OperationSelector } from './components/OperationSelector';
import { useCalculator } from './hooks/useCalculator';
import './App.css';

function App() {
  const calc = useCalculator();

  return (
    <main className="app">
      <section className="card">
        <h1 className="title">Calculator</h1>

        <Display result={calc.result} error={calc.error} isLoading={calc.isLoading} />

        <OperationSelector
          value={calc.operation}
          onChange={calc.changeOperation}
          disabled={calc.isLoading}
        />

        <div className="operands">
          <OperandInput
            id="operand-a"
            label={calc.needsSecondOperand ? 'First number' : 'Number'}
            value={calc.a}
            onChange={calc.setA}
            disabled={calc.isLoading}
          />
          {calc.needsSecondOperand && (
            <OperandInput
              id="operand-b"
              label="Second number"
              value={calc.b}
              onChange={calc.setB}
              disabled={calc.isLoading}
            />
          )}
        </div>

        <div className="actions">
          <button
            type="button"
            className="btn btn--primary"
            onClick={calc.submit}
            disabled={calc.isLoading}
          >
            Calculate
          </button>
          <button type="button" className="btn" onClick={calc.clear} disabled={calc.isLoading}>
            Clear
          </button>
        </div>
      </section>
    </main>
  );
}

export default App;