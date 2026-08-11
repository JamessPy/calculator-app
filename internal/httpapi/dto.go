package httpapi

// calculateRequest is the body of POST /api/v1/calculate.
type calculateRequest struct {
	Operation string   `json:"operation"`
	A         *float64 `json:"a"`
	B         *float64 `json:"b"`
}

// calculateResponse is the body of a successful calculation.
type calculateResponse struct {
	Operation string   `json:"operation"`
	A         float64  `json:"a"`
	B         *float64 `json:"b,omitempty"`
	Result    float64  `json:"result"`
}

// errorResponse is the envelope for every error the API returns.
type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`    // stable, machine-readable
	Message string `json:"message"` // human-readable, may change
}
