//go:build integration

package opensearch

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/example/ase/internal/config"
)

// TestLiveOpenSearch_Search requires OPENSEARCH_URLS and OPENSEARCH_INDEX (and reachable cluster).
func TestLiveOpenSearch_Search(t *testing.T) {
	if strings.TrimSpace(os.Getenv("OPENSEARCH_URLS")) == "" {
		t.Skip("set OPENSEARCH_URLS to run integration tests")
	}
	cfg := config.Load()
	if len(cfg.OpenSearchURLs) == 0 || cfg.OpenSearchIndex == "" {
		t.Skip("OPENSEARCH_URLS or OPENSEARCH_INDEX empty after config.Load")
	}
	repo, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestDeadline)
	defer cancel()
	_, err = repo.Search(ctx, "integration_probe")
	if err != nil {
		t.Fatal(err)
	}
}
