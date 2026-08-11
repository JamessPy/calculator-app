package calculator

import "testing"

func TestOperation_IsValid(t *testing.T) {
	tests := []struct {
		name string
		op   Operation
		want bool
	}{
		{name: "add", op: OpAdd, want: true},
		{name: "subtract", op: OpSubtract, want: true},
		{name: "multiply", op: OpMultiply, want: true},
		{name: "divide", op: OpDivide, want: true},
		{name: "power", op: OpPower, want: true},
		{name: "sqrt", op: OpSqrt, want: true},
		{name: "percentage", op: OpPercentage, want: true},

		{name: "empty string", op: Operation(""), want: false},
		{name: "unknown operation", op: Operation("modulo"), want: false},

		// The domain is strict; normalisation is the transport layer's job.
		{name: "uppercase is rejected", op: Operation("ADD"), want: false},
		{name: "surrounding spaces are rejected", op: Operation(" add "), want: false},

		// Regression guard: the wire value must stay lowercase, not the Go
		// identifier. Using "OpAdd" as the constant value would break the
		// API contract while leaving every domain test green.
		{name: "go identifier name is rejected", op: Operation("OpAdd"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.IsValid(); got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}

func TestOperation_IsUnary(t *testing.T) {
	tests := []struct {
		name string
		op   Operation
		want bool
	}{
		{name: "sqrt is unary", op: OpSqrt, want: true},
		{name: "add is binary", op: OpAdd, want: false},
		{name: "divide is binary", op: OpDivide, want: false},
		{name: "power is binary", op: OpPower, want: false},
		{name: "unknown is not unary", op: Operation("modulo"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.IsUnary(); got != tt.want {
				t.Errorf("IsUnary(%q) = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}
