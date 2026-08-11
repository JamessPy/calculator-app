package calculator

import (
	"errors"
	"math"
	"testing"
)

// ptr is a helper function to create a pointer to a float64 value.
func ptr(f float64) *float64 {
	return &f
}

func TestCalculate(t *testing.T) {

	tests := []struct {
		name    string
		op      Operation
		a       float64
		b       *float64
		want    float64
		wantErr error
	}{
		// binary operations
		{name: "add", op: OpAdd, a: 2, b: ptr(3), want: 5},
		{name: "add negative", op: OpAdd, a: -5, b: ptr(3), want: -2},
		{name: "subtract", op: OpSubtract, a: 10, b: ptr(4), want: 6},
		{name: "multiply", op: OpMultiply, a: 6, b: ptr(7), want: 42},
		{name: "multiply by zero", op: OpMultiply, a: 12345, b: ptr(0), want: 0},
		{name: "divide", op: OpDivide, a: 10, b: ptr(4), want: 2.5},
		{name: "divide by negative", op: OpDivide, a: 10, b: ptr(-2), want: -5},
		{name: "power", op: OpPower, a: 2, b: ptr(10), want: 1024},
		{name: "power negative exponent", op: OpPower, a: 2, b: ptr(-2), want: 0.25},
		{name: "power zero exponent", op: OpPower, a: 5, b: ptr(0), want: 1},
		{name: "percentage", op: OpPercentage, a: 200, b: ptr(15), want: 30},

		// unary operations
		{name: "sqrt", op: OpSqrt, a: 144, b: nil, want: 12},
		{name: "sqrt zero", op: OpSqrt, a: 0, b: nil, want: 0},
		{name: "sqrt ignores second operand", op: OpSqrt, a: 144, b: ptr(999), want: 12},

		// error cases
		{name: "divide by zero", op: OpDivide, a: 1, b: ptr(0), wantErr: ErrDivisionByZero},
		{name: "zero divided by zero", op: OpDivide, a: 0, b: ptr(0), wantErr: ErrDivisionByZero},
		{name: "sqrt of negative", op: OpSqrt, a: -9, b: nil, wantErr: ErrNegativeSqrt},
		{name: "unsupported operation", op: Operation("modulo"), a: 1, b: ptr(2), wantErr: ErrUnsupportedOperation},
		{name: "empty operation", op: Operation(""), a: 1, b: ptr(2), wantErr: ErrUnsupportedOperation},
		{name: "missing second operand", op: OpAdd, a: 1, b: nil, wantErr: ErrOperandRequired},
		{name: "multiplication overflow", op: OpMultiply, a: 1e308, b: ptr(10), wantErr: ErrResultNotFinite},
		{name: "zero to negative power", op: OpPower, a: 0, b: ptr(-1), wantErr: ErrResultNotFinite},
		{name: "NaN operand", op: OpAdd, a: math.NaN(), b: ptr(1), wantErr: ErrOperandNotFinite},
		{name: "Inf operand", op: OpAdd, a: math.Inf(1), b: ptr(1), wantErr: ErrOperandNotFinite},
		{name: "NaN second operand", op: OpAdd, a: 1, b: ptr(math.NaN()), wantErr: ErrOperandNotFinite},
	}

	svc := NewService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Calculate(tt.op, tt.a, tt.b)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// The default branches in unary() and binary() are unreachable through
// Calculate(), which validates the operation before dispatching.
func TestService_DispatchGuards(t *testing.T) {
	svc := NewService()

	t.Run("unary rejects a binary operation", func(t *testing.T) {
		_, err := svc.unary(OpAdd, 5)
		if !errors.Is(err, ErrUnsupportedOperation) {
			t.Errorf("got %v, want %v", err, ErrUnsupportedOperation)
		}
	})

	t.Run("binary rejects a unary operation", func(t *testing.T) {
		_, err := svc.binary(OpSqrt, 4, 2)
		if !errors.Is(err, ErrUnsupportedOperation) {
			t.Errorf("got %v, want %v", err, ErrUnsupportedOperation)
		}
	})
}
