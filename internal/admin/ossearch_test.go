package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ase/internal/admin"
	"github.com/example/ase/internal/config"
)

func TestIndexSearchRaw_roundTrip(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/my-index/_search") {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hits":{"total":{"value":0},"hits":[]}}`))
	}))
	defer ts.Close()

	cfg := config.Config{
		OpenSearchURLs:  []string{ts.URL},
		OpenSearchIndex: "my-index",
	}
	ctx := context.Background()
	b, code, err := admin.IndexSearchRaw(ctx, cfg, []byte(`{"query":{"match_all":{}},"size":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	if !strings.Contains(string(b), `"hits"`) {
		t.Fatalf("unexpected body: %s", b)
	}
}
