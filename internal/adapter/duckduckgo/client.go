// Package duckduckgo implements port.SearchProvider via DuckDuckGo HTML search (no API key).
// It POSTs to https://html.duckduckgo.com/html/ and parses organic result links; subject to HTML changes and fair-use limits.
package duckduckgo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/port"
)

const defaultBase = "https://html.duckduckgo.com/html"

// Client fetches DuckDuckGo HTML SERP and extracts result rows.
type Client struct {
	HTTP       *http.Client
	BaseURL    string
	MaxResults int
	UserAgent  string
}

// NewFromConfig returns nil when DUCKDUCKGO_ENABLED is false.
func NewFromConfig(cfg config.Config) *Client {
	if !cfg.DuckDuckGoEnabled {
		return nil
	}
	max := cfg.DuckDuckGoMaxResults
	if max < 1 {
		max = 10
	}
	if max > 25 {
		max = 25
	}
	base := strings.TrimSpace(cfg.DuckDuckGoBaseURL)
	if base == "" {
		base = defaultBase
	}
	ms := cfg.DuckDuckGoTimeoutMs
	if ms <= 0 {
		ms = 15000
	}
	ua := strings.TrimSpace(cfg.DuckDuckGoUserAgent)
	if ua == "" {
		ua = "Mozilla/5.0 (compatible; ASE/1.0; +https://github.com/example/ase)"
	}
	return &Client{
		HTTP: &http.Client{
			Timeout: time.Duration(ms) * time.Millisecond,
		},
		BaseURL:    strings.TrimSuffix(base, "/"),
		MaxResults: max,
		UserAgent:  ua,
	}
}

// Search implements port.SearchProvider.
func (c *Client) Search(ctx context.Context, query string) port.ProviderResult {
	q := strings.TrimSpace(query)
	if q == "" {
		return port.ProviderResult{Items: nil}
	}
	form := url.Values{}
	form.Set("q", q)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return port.ProviderResult{Err: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.http().Do(req)
	if err != nil {
		return port.ProviderResult{Err: fmt.Errorf("duckduckgo: request: %w", err)}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 3<<20))
	if err != nil {
		return port.ProviderResult{Err: fmt.Errorf("duckduckgo: read body: %w", err)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return port.ProviderResult{Err: fmt.Errorf("duckduckgo: http %d: %s", resp.StatusCode, truncateErr(body))}
	}
	items, err := parseResultsHTML(strings.NewReader(string(body)), c.MaxResults)
	if err != nil {
		return port.ProviderResult{Err: err}
	}
	return port.ProviderResult{Items: items}
}

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func parseResultsHTML(r io.Reader, max int) ([]port.ProviderItem, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("duckduckgo: parse html: %w", err)
	}
	var items []port.ProviderItem
	doc.Find(".result").Each(func(_ int, row *goquery.Selection) {
		if len(items) >= max {
			return
		}
		a := row.Find("a.result__a").First()
		if a.Length() == 0 {
			return
		}
		href, _ := a.Attr("href")
		href = normalizeResultURL(href)
		if href == "" {
			return
		}
		title := strings.TrimSpace(a.Text())
		snippet := strings.TrimSpace(row.Find("a.result__snippet").First().Text())
		if snippet == "" {
			snippet = strings.TrimSpace(row.Find(".result__snippet").First().Text())
		}
		items = append(items, port.ProviderItem{
			URL:     href,
			Title:   title,
			Snippet: snippet,
		})
	})
	return items, nil
}

func normalizeResultURL(href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	}
	u, err := url.Parse(href)
	if err != nil {
		return href
	}
	if u.Host == "" {
		return href
	}
	// DuckDuckGo redirect wrapper: /l/?uddg=https%3A%2F%2F...
	if q := u.Query().Get("uddg"); q != "" {
		if decoded, err := url.QueryUnescape(q); err == nil && decoded != "" {
			return decoded
		}
	}
	return href
}

func truncateErr(b []byte) string {
	const max = 256
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
