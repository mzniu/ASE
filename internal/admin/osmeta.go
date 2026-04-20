package admin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/example/ase/internal/config"
)

// CatIndicesJSON returns the raw JSON body from GET /_cat/indices?format=json against the first OpenSearch URL.
func CatIndicesJSON(ctx context.Context, cfg config.Config) ([]byte, int, error) {
	if len(cfg.OpenSearchURLs) == 0 {
		return []byte("[]"), http.StatusOK, nil
	}
	base := strings.TrimRight(cfg.OpenSearchURLs[0], "/")
	u := base + "/_cat/indices?format=json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	if cfg.OpenSearchUser != "" {
		req.SetBasicAuth(cfg.OpenSearchUser, cfg.OpenSearchPassword)
	}
	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return b, resp.StatusCode, fmt.Errorf("opensearch: status %d", resp.StatusCode)
	}
	return b, resp.StatusCode, nil
}
