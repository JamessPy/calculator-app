package calculator

import (
	"errors"
)

// Supported operations
type Operation string

const (
	OpAdd        Operation = "add"
	OpSubtract   Operation = "subtract"
	OpMultiply   Operation = "multiply"
	OpDivide     Operation = "divide"
	OpPower      Operation = "power"
	OpSqrt       Operation = "sqrt"
	OpPercentage Operation = "percentage"
)

// Sentinel error for unsupported operations
var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrDivisionByZero       = errors.New("division by zero")
	ErrNegativeSqrt         = errors.New("cannot calculate square root of a negative number")
	ErrOperandRequired      = errors.New("operation requires at least one operand")
	ErrOperandNotFinite     = errors.New("operand is not a finite number")
	ErrResultNotFinite      = errors.New("result is not a finite number")
)

// Is valid checks if the operation is supported
func (op Operation) IsValid() bool {
	switch op {
	case OpAdd, OpSubtract, OpMultiply, OpDivide, OpPower, OpSqrt, OpPercentage:
		return true
	default:
		return false
	}
}

// IsUnary checks if the operation is unary (requires only one operand)
func (op Operation) IsUnary() bool {
	return op == OpSqrt
}
