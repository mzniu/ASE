package admin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/ase/internal/config"
)

const maxSearchResponseBytes = 1 << 22 // 4 MiB cap for admin browse responses

// IndexSearchRaw POSTs JSON to {index}/_search on the first configured OpenSearch URL (same auth as CatIndicesJSON).
func IndexSearchRaw(ctx context.Context, cfg config.Config, body []byte) ([]byte, int, error) {
	if len(cfg.OpenSearchURLs) == 0 || strings.TrimSpace(cfg.OpenSearchIndex) == "" {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("opensearch not configured")
	}
	if len(body) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empty search body")
	}
	base := strings.TrimRight(cfg.OpenSearchURLs[0], "/")
	idx := url.PathEscape(cfg.OpenSearchIndex)
	u := base + "/" + idx + "/_search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.OpenSearchUser != "" {
		req.SetBasicAuth(cfg.OpenSearchUser, cfg.OpenSearchPassword)
	}
	c := &http.Client{Timeout: 25 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxSearchResponseBytes))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return b, resp.StatusCode, fmt.Errorf("opensearch: status %d", resp.StatusCode)
	}
	return b, resp.StatusCode, nil
}
