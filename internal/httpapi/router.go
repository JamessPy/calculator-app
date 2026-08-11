package httpapi

import (
	"log/slog"
	"net/http"
)

//NewRouter wires up the HTTP routes for the service

func NewRouter(calc Calculator, logger *slog.Logger, allowedOrigins []string) http.Handler {
	h := NewHandler(calc, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.Health)
	mux.HandleFunc("POST /api/v1/calculate", h.Calculate)

	return CORS(allowedOrigins)(mux)
}
