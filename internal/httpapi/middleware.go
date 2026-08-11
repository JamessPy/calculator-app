package httpapi

import (
	"net/http"
	"slices"
)

// CORS allows the frontend, served from a different port, to call this API.
// The browser treats a different port as a different origin and blocks the
// response unless these headers are present.
//
// Origins are matched against an explicit allow-list rather than reflecting
// the request's Origin header, which would be equivalent to allowing all.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && slices.Contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}

			// Preflight: the browser asks permission before the real request.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
