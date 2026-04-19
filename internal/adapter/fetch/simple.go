package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/ase/internal/port"
	"golang.org/x/sync/errgroup"
)

const maxBodyBytes = 2 << 20 // 2 MiB

// SimpleConfig tunes the HTTP client used by Simple.
type SimpleConfig struct {
	PerURLTimeout time.Duration
	UserAgent     string
	// Concurrency is the max number of in-flight HTTPS GETs (default 4 if <= 0).
	Concurrency int
}

// Simple is a PageFetcher: HTTPS GET, readability + HTML→Markdown, bounded concurrent requests.
type Simple struct {
	client      *http.Client
	ua          string
	concurrency int
}

// NewSimple returns a PageFetcher backed by net/http. perURLTimeout must be > 0.
func NewSimple(c SimpleConfig) *Simple {
	if c.PerURLTimeout <= 0 {
		c.PerURLTimeout = 8 * time.Second
	}
	conc := c.Concurrency
	if conc <= 0 {
		conc = 4
	}
	ua := strings.TrimSpace(c.UserAgent)
	if ua == "" {
		ua = "ASE/1.0 (+https://github.com/example/ase)"
	}
	return &Simple{
		client:      &http.Client{Timeout: c.PerURLTimeout},
		ua:          ua,
		concurrency: conc,
	}
}

// FetchPlainText implements port.PageFetcher.
func (s *Simple) FetchPlainText(ctx context.Context, urls []string, limit int) []port.FetchedPage {
	if limit <= 0 {
		return nil
	}
	toFetch := collectFetchableURLs(urls, limit)
	if len(toFetch) == 0 {
		return nil
	}
	conc := s.concurrency
	if conc < 1 {
		conc = 4
	}
	results := make([]port.FetchedPage, len(toFetch))
	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, conc)
	for i, u := range toFetch {
		i, u := i, u
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()
			var text string
			if ctx.Err() == nil {
				text = s.fetchOneMarkdown(ctx, u)
			}
			results[i] = port.FetchedPage{URL: u, Text: text}
			return nil
		})
	}
	_ = g.Wait()
	return results
}

// collectFetchableURLs returns up to limit distinct http(s) URLs (Baidu organic links are often http://).
func collectFetchableURLs(urls []string, limit int) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, raw := range urls {
		if len(out) >= limit {
			break
		}
		u := strings.TrimSpace(raw)
		if u == "" {
			continue
		}
		if _, dup := seen[u]; dup {
			continue
		}
		pu, err := url.Parse(u)
		if err != nil || (pu.Scheme != "http" && pu.Scheme != "https") || pu.Host == "" {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func (s *Simple) fetchOneMarkdown(ctx context.Context, pageURL string) string {
	body, err := s.fetchBody(ctx, pageURL)
	if err != nil || len(body) == 0 {
		return ""
	}
	return HTMLMainToMarkdown(body, pageURL)
}

func (s *Simple) fetchBody(ctx context.Context, pageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBodyBytes {
		body = body[:maxBodyBytes]
	}
	return body, nil
}
