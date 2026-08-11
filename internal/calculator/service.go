package calculator

import "math"

// Service operates arithmetic operations
type Service struct{}

// NewService creates a new instance of Service
func NewService() *Service {
	return &Service{}
}

// Calculate, It takes an operation and two operands, and returns the result of the operation
func (s *Service) Calculate(op Operation, a float64, b *float64) (float64, error) {
	if !op.IsValid() {
		return 0, ErrUnsupportedOperation
	}
	if !isFinite(a) {
		return 0, ErrOperandNotFinite
	}
	if b != nil && !isFinite(*b) {
		return 0, ErrOperandNotFinite
	}
	if op.IsUnary() {
		return s.unary(op, a)
	}
	// For binary options, 'b' is mandatory.
	if b == nil {
		return 0, ErrOperandRequired
	}
	return s.binary(op, a, *b)
}

// unary, it operates 1 operand and returns the result of the operation
func (s *Service) unary(op Operation, a float64) (float64, error) {
	switch op {
	case OpSqrt:
		if a < 0 {
			return 0, ErrNegativeSqrt
		}
		return math.Sqrt(a), nil
	default:
		return 0, ErrUnsupportedOperation
	}
}

// binary, it operates 2 operands and returns the result of the operation
func (s *Service) binary(op Operation, a, b float64) (float64, error) {
	var result float64
	switch op {
	case OpAdd:
		result = a + b
	case OpSubtract:
		result = a - b
	case OpMultiply:
		result = a * b
	case OpDivide:
		if b == 0 {
			return 0, ErrDivisionByZero
		}
		result = a / b
	case OpPower:
		result = math.Pow(a, b)
	case OpPercentage:
		result = a * b / 100
	default:
		return 0, ErrUnsupportedOperation
	}

	// check nan and inf
	if !isFinite(result) {
		return 0, ErrResultNotFinite
	}
	return result, nil
}

// isFinite checks if a float64 value is finite (not NaN or Inf)
func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}
