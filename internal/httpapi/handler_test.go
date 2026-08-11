package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JamessPy/calculator-app/internal/calculator"
)

// discardLogger keeps test output free of log lines.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestServer wires the router with the real domain service, so the test
// exercises the full transport + domain path.
var testOrigins = []string{"http://localhost:5173"}

func newTestServer() http.Handler {
	return NewRouter(calculator.NewService(), discardLogger(), testOrigins)
}

// doRequest sends a request through the router without opening a socket.
func doRequest(srv http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestCalculate_Success(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantOp     string
		wantResult float64
		wantB      bool // should the response contain a "b" field?
	}{
		{name: "add", body: `{"operation":"add","a":2,"b":3}`, wantOp: "add", wantResult: 5, wantB: true},
		{name: "subtract", body: `{"operation":"subtract","a":10,"b":4}`, wantOp: "subtract", wantResult: 6, wantB: true},
		{name: "multiply", body: `{"operation":"multiply","a":6,"b":7}`, wantOp: "multiply", wantResult: 42, wantB: true},
		{name: "divide", body: `{"operation":"divide","a":10,"b":4}`, wantOp: "divide", wantResult: 2.5, wantB: true},
		{name: "power", body: `{"operation":"power","a":2,"b":10}`, wantOp: "power", wantResult: 1024, wantB: true},
		{name: "percentage", body: `{"operation":"percentage","a":200,"b":15}`, wantOp: "percentage", wantResult: 30, wantB: true},
		{name: "sqrt omits b", body: `{"operation":"sqrt","a":144}`, wantOp: "sqrt", wantResult: 12, wantB: false},
		{name: "zero operand is accepted", body: `{"operation":"add","a":0,"b":5}`, wantOp: "add", wantResult: 5, wantB: true},

		// normalisation is a transport-layer concern
		{name: "uppercase is normalised", body: `{"operation":"ADD","a":2,"b":3}`, wantOp: "add", wantResult: 5, wantB: true},
		{name: "whitespace is trimmed", body: `{"operation":" add ","a":2,"b":3}`, wantOp: "add", wantResult: 5, wantB: true},
	}

	srv := newTestServer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(srv, http.MethodPost, "/api/v1/calculate", tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			var resp calculateResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("cannot decode response: %v", err)
			}
			if resp.Operation != tt.wantOp {
				t.Errorf("operation = %q, want %q", resp.Operation, tt.wantOp)
			}
			if resp.Result != tt.wantResult {
				t.Errorf("result = %v, want %v", resp.Result, tt.wantResult)
			}

			// "b" must be absent for unary operations, present otherwise.
			_, hasB := rawFields(t, rec.Body.Bytes())["b"]
			if hasB != tt.wantB {
				t.Errorf("response has 'b' = %v, want %v; body = %s", hasB, tt.wantB, rec.Body.String())
			}
		})
	}
}

// rawFields decodes the body into a generic map so the test can assert on
// which JSON keys are present, not just on the decoded struct.
func rawFields(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("cannot decode body as map: %v", err)
	}
	return m
}
func TestCalculate_Errors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		// 422: the request is well-formed, the mathematics is not defined
		{name: "divide by zero", body: `{"operation":"divide","a":1,"b":0}`,
			wantStatus: http.StatusUnprocessableEntity, wantCode: codeDivisionByZero},
		{name: "sqrt of negative", body: `{"operation":"sqrt","a":-9}`,
			wantStatus: http.StatusUnprocessableEntity, wantCode: codeNegativeSquareRoot},
		{name: "overflow", body: `{"operation":"power","a":1e308,"b":2}`,
			wantStatus: http.StatusUnprocessableEntity, wantCode: codeResultNotFinite},

		// 400: the request itself is wrong
		{name: "unknown operation", body: `{"operation":"modulo","a":1,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: codeUnsupportedOperation},
		{name: "missing second operand", body: `{"operation":"add","a":1}`,
			wantStatus: http.StatusBadRequest, wantCode: codeOperandRequired},
		{name: "missing operation", body: `{"a":1,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: codeValidationFailed},
		{name: "empty operation", body: `{"operation":"","a":1,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: codeValidationFailed},
		{name: "missing first operand", body: `{"operation":"add","b":1}`,
			wantStatus: http.StatusBadRequest, wantCode: codeValidationFailed},
		{name: "operand is a string", body: `{"operation":"add","a":"two","b":3}`,
			wantStatus: http.StatusBadRequest, wantCode: codeInvalidJSON},
		{name: "operand out of float64 range", body: `{"operation":"add","a":1e400,"b":3}`,
			wantStatus: http.StatusBadRequest, wantCode: codeInvalidJSON},
		{name: "malformed JSON", body: `{"operation":"add",`,
			wantStatus: http.StatusBadRequest, wantCode: codeInvalidJSON},
		{name: "empty body", body: ``,
			wantStatus: http.StatusBadRequest, wantCode: codeInvalidJSON},
	}

	srv := newTestServer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(srv, http.MethodPost, "/api/v1/calculate", tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			var resp errorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("cannot decode error response: %v", err)
			}
			if resp.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
			}
			if resp.Error.Message == "" {
				t.Error("error message must not be empty")
			}
		})
	}
}

func TestRouting(t *testing.T) {
	srv := newTestServer()

	t.Run("health check", func(t *testing.T) {
		rec := doRequest(srv, http.MethodGet, "/healthz", "")
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("wrong method is rejected", func(t *testing.T) {
		rec := doRequest(srv, http.MethodGet, "/api/v1/calculate", "")
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("unknown path is rejected", func(t *testing.T) {
		rec := doRequest(srv, http.MethodPost, "/api/v1/nope", `{}`)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})
}

// stubCalculator returns a fixed error, so the handler's unexpected-error
// path can be exercised. The real service never produces such an error,
// which is precisely why a stub is needed here.
type stubCalculator struct{ err error }

func (s stubCalculator) Calculate(calculator.Operation, float64, *float64) (float64, error) {
	return 0, s.err
}

func TestCalculate_UnexpectedDomainError(t *testing.T) {
	srv := NewRouter(stubCalculator{err: errors.New("database on fire")}, discardLogger(), testOrigins)

	rec := doRequest(srv, http.MethodPost, "/api/v1/calculate", `{"operation":"add","a":1,"b":2}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("cannot decode error response: %v", err)
	}
	if resp.Error.Code != codeInternal {
		t.Errorf("code = %q, want %q", resp.Error.Code, codeInternal)
	}
	// Internal detail must never reach the client.
	if strings.Contains(rec.Body.String(), "database on fire") {
		t.Error("internal error detail leaked into the response")
	}
}

// / mapDomainError is tested directly for branches that cannot be reached
// through HTTP: JSON has no NaN/Infinity literal, so a non-finite operand
// never survives decoding. The mapping is still part of the contract.
func TestMapDomainError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"unsupported operation", calculator.ErrUnsupportedOperation, http.StatusBadRequest, codeUnsupportedOperation},
		{"operand required", calculator.ErrOperandRequired, http.StatusBadRequest, codeOperandRequired},
		{"operand not finite", calculator.ErrOperandNotFinite, http.StatusBadRequest, codeOperandNotFinite},
		{"division by zero", calculator.ErrDivisionByZero, http.StatusUnprocessableEntity, codeDivisionByZero},
		{"negative square root", calculator.ErrNegativeSqrt, http.StatusUnprocessableEntity, codeNegativeSquareRoot},
		{"result not finite", calculator.ErrResultNotFinite, http.StatusUnprocessableEntity, codeResultNotFinite},
		{"unknown error", errors.New("something else"), http.StatusInternalServerError, codeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := mapDomainError(tt.err)
			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
			if body.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", body.Code, tt.wantCode)
			}
		})
	}
}

// Wrapped errors must still map correctly: errors.Is walks the chain,
// so the domain is free to add context without breaking the HTTP contract.
func TestMapDomainError_WrappedError(t *testing.T) {
	wrapped := fmt.Errorf("binary operation failed: %w", calculator.ErrDivisionByZero)

	status, body := mapDomainError(wrapped)

	if status != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", status)
	}
	if body.Code != codeDivisionByZero {
		t.Errorf("code = %q, want %q", body.Code, codeDivisionByZero)
	}
}

func TestNewHandler_NilLoggerFallsBackToDefault(t *testing.T) {
	h := NewHandler(calculator.NewService(), nil)
	if h.logger == nil {
		t.Error("logger must not be nil")
	}
}

func TestCORS(t *testing.T) {
	srv := newTestServer()

	t.Run("allowed origin receives CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate",
			strings.NewReader(`{"operation":"add","a":1,"b":2}`))
		req.Header.Set("Origin", testOrigins[0])
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testOrigins[0] {
			t.Errorf("Allow-Origin = %q, want %q", got, testOrigins[0])
		}
		if got := rec.Header().Get("Vary"); got != "Origin" {
			t.Errorf("Vary = %q, want Origin", got)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("disallowed origin receives no CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate",
			strings.NewReader(`{"operation":"add","a":1,"b":2}`))
		req.Header.Set("Origin", "http://evil.example.com")
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Allow-Origin = %q, want empty", got)
		}
		// The request still succeeds server-side; CORS is a browser policy,
		// not server-side authorisation. The browser withholds the response
		// from the calling script when the header is absent.
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("request without an Origin header is untouched", func(t *testing.T) {
		// Non-browser clients (curl, server-to-server) send no Origin.
		rec := doRequest(srv, http.MethodPost, "/api/v1/calculate", `{"operation":"add","a":1,"b":2}`)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Allow-Origin = %q, want empty", got)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("preflight is answered without reaching the handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/calculate", nil)
		req.Header.Set("Origin", testOrigins[0])
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want 204", rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
			t.Error("Allow-Methods header is missing")
		}
		if rec.Body.Len() != 0 {
			t.Errorf("preflight response has a body: %q", rec.Body.String())
		}
	})
}
