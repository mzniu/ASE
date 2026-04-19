package httpx

import (
	"context"
	"errors"
	"net/http"

	"github.com/example/ase/internal/orchestrator"
)

// WriteSearchFailure maps orchestrator / dependency errors to Problem Details (REQ-F-010, SEARCH_API_V1 §3.5).
// Deadline exceeded → 504; bad request (providers) → 400; otherwise 503.
func WriteSearchFailure(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, orchestrator.ErrBadRequest):
		WriteProblem(w, http.StatusBadRequest, "validation error", err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		WriteProblem(w, http.StatusGatewayTimeout, "gateway timeout",
			"search deadline exceeded before a usable result was available")
	case errors.Is(err, context.Canceled):
		WriteProblem(w, http.StatusServiceUnavailable, "dependency unavailable",
			"search canceled or upstream closed")
	default:
		WriteProblem(w, http.StatusServiceUnavailable, "dependency unavailable", err.Error())
	}
}
