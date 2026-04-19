package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/example/ase/internal/orchestrator"
)

func TestWriteSearchFailure_statusAndTitle(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantTitle  string
	}{
		{"bad_request", orchestrator.ErrBadRequest, 400, "validation error"},
		{"deadline", context.DeadlineExceeded, 504, "gateway timeout"},
		{"canceled", context.Canceled, 503, "dependency unavailable"},
		{"wrapped_deadline", fmt.Errorf("wrap: %w", context.DeadlineExceeded), 504, "gateway timeout"},
		{"generic", errors.New("upstream down"), 503, "dependency unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()
			WriteSearchFailure(rr, tc.err)
			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			var p ProblemDetail
			if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
				t.Fatal(err)
			}
			if p.Title != tc.wantTitle {
				t.Fatalf("title = %q, want %q", p.Title, tc.wantTitle)
			}
			if p.Status != tc.wantStatus {
				t.Fatalf("problem status field = %d, want %d", p.Status, tc.wantStatus)
			}
		})
	}
}
