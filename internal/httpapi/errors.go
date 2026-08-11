package httpapi

import (
	"errors"
	"net/http"

	"github.com/JamessPy/calculator-app/internal/calculator"
)

const (
	codeInvalidJSON          = "INVALID_JSON"
	codeValidationFailed     = "VALIDATION_FAILED"
	codeUnsupportedOperation = "UNSUPPORTED_OPERATION"
	codeDivisionByZero       = "DIVISION_BY_ZERO"
	codeNegativeSquareRoot   = "NEGATIVE_SQUARE_ROOT"
	codeOperandRequired      = "OPERAND_REQUIRED"
	codeOperandNotFinite     = "OPERAND_NOT_FINITE"
	codeResultNotFinite      = "RESULT_NOT_FINITE"
	codeInternal             = "INTERNAL_ERROR"
)

// mapDomainError translates a domain error into an HTTP status and an API error body.
//
//	400 Bad Request          — the request is malformed or nonsensical
//	422 Unprocessable Entity — the request is well-formed but mathematically undefined
func mapDomainError(err error) (int, errorBody) {
	switch {
	case errors.Is(err, calculator.ErrUnsupportedOperation):
		return http.StatusBadRequest, errorBody{
			Code:    codeUnsupportedOperation,
			Message: "unsupported operation; supported: add, subtract, multiply, divide, power, sqrt, percentage",
		}
	case errors.Is(err, calculator.ErrOperandRequired):
		return http.StatusBadRequest, errorBody{
			Code:    codeOperandRequired,
			Message: "field 'b' is required for this operation",
		}
	case errors.Is(err, calculator.ErrOperandNotFinite):
		return http.StatusBadRequest, errorBody{
			Code:    codeOperandNotFinite,
			Message: "operands must be finite numbers",
		}
	case errors.Is(err, calculator.ErrDivisionByZero):
		return http.StatusUnprocessableEntity, errorBody{
			Code:    codeDivisionByZero,
			Message: "division by zero is undefined",
		}
	case errors.Is(err, calculator.ErrNegativeSqrt):
		return http.StatusUnprocessableEntity, errorBody{
			Code:    codeNegativeSquareRoot,
			Message: "square root of a negative number is undefined in real numbers",
		}
	case errors.Is(err, calculator.ErrResultNotFinite):
		return http.StatusUnprocessableEntity, errorBody{
			Code:    codeResultNotFinite,
			Message: "result exceeds the representable numeric range",
		}
	default:
		return http.StatusInternalServerError, errorBody{
			Code:    codeInternal,
			Message: "internal server error",
		}
	}
}
