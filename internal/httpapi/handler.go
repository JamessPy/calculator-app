// Package httpapi exposes the calculator over HTTP.
// It is the only package that knows about HTTP, JSON and status codes.
package httpapi

import (
	"encoding/json"
	"github.com/JamessPy/calculator-app/internal/calculator"
	"log/slog"
	"net/http"
	"strings"
)

// Calculator is the behaviour this package needs from the domain layer.
//
// Go idiom: the consuming package declares the interface, not the producing
// one. internal/calculator is unaware of httpapi, and tests can substitute
// a stub without touching the domain.
type Calculator interface {
	Calculate(op calculator.Operation, a float64, b *float64) (float64, error)
}

// Handler serves the HTTP endpoints of the service.
type Handler struct {
	calc   Calculator
	logger *slog.Logger
}

// NewHandler creates a new Handler.
func NewHandler(calc Calculator, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{calc: calc, logger: logger}
}

// Health responds to GET /healthz.
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Calculate responds to POST /api/v1/calculate.
// Calculate responds to POST /api/v1/calculate.
func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req calculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, errorBody{
			Code:    codeInvalidJSON,
			Message: "request body is not valid JSON",
		})
		return
	}

	if strings.TrimSpace(req.Operation) == "" {
		h.writeError(w, http.StatusBadRequest, errorBody{
			Code:    codeValidationFailed,
			Message: "field 'operation' is required",
		})
		return
	}
	if req.A == nil {
		h.writeError(w, http.StatusBadRequest, errorBody{
			Code:    codeValidationFailed,
			Message: "field 'a' is required and must be a number",
		})
		return
	}

	// Normalisation belongs to the transport layer: the domain stays strict
	// while the API stays forgiving about case and surrounding whitespace.
	op := calculator.Operation(strings.ToLower(strings.TrimSpace(req.Operation)))

	result, err := h.calc.Calculate(op, *req.A, req.B)
	if err != nil {
		status, body := mapDomainError(err)
		if status == http.StatusInternalServerError {
			h.logger.Error("calculation failed unexpectedly", "operation", op, "err", err)
		}
		h.writeError(w, status, body)
		return
	}

	resp := calculateResponse{
		Operation: string(op),
		A:         *req.A,
		Result:    result,
	}
	if !op.IsUnary() {
		resp.B = req.B
	}
	h.writeJSON(w, http.StatusOK, resp)
}

// writeJSON serialises payload as JSON and writes it with the given status.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The status code and headers are already sent, so the response
		// cannot be corrected at this point. Logging is all we can do.
		h.logger.Error("failed to encode response", "err", err)
	}
}

// writeError writes an error body in the standard envelope.
func (h *Handler) writeError(w http.ResponseWriter, status int, body errorBody) {
	h.writeJSON(w, status, errorResponse{Error: body})
}
