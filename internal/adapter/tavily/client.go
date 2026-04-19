// Package tavily implements port.SearchProvider against the Tavily Search API.
package tavily

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/port"
)

// Client calls POST /search on the Tavily API (https://docs.tavily.com).
type Client struct {
	HTTP        *http.Client
	APIKey      string
	BaseURL     string
	MaxResults  int
	SearchDepth string
	ProjectID   string
}

// NewFromConfig returns nil if TAVILY_API_KEY is unset.
func NewFromConfig(cfg config.Config) *Client {
	if strings.TrimSpace(cfg.TavilyAPIKey) == "" {
		return nil
	}
	max := cfg.TavilyMaxResults
	if max < 1 {
		max = 10
	}
	if max > 20 {
		max = 20
	}
	depth := cfg.TavilySearchDepth
	if depth == "" {
		depth = "basic"
	}
	return &Client{
		HTTP:        http.DefaultClient,
		APIKey:      strings.TrimSpace(cfg.TavilyAPIKey),
		BaseURL:     strings.TrimSuffix(strings.TrimSpace(cfg.TavilyBaseURL), "/"),
		MaxResults:  max,
		SearchDepth: depth,
		ProjectID:   strings.TrimSpace(cfg.TavilyProjectID),
	}
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

type searchRequest struct {
	Query             string `json:"query"`
	SearchDepth       string `json:"search_depth"`
	MaxResults        int    `json:"max_results"`
	IncludeAnswer     bool   `json:"include_answer"`
	IncludeRawContent bool   `json:"include_raw_content"`
}

type searchResponse struct {
	Results []resultItem `json:"results"`
}

type resultItem struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Content    string `json:"content"`
	RawContent string `json:"raw_content"`
}

// Search implements port.SearchProvider.
func (c *Client) Search(ctx context.Context, query string) port.ProviderResult {
	endpoint := c.BaseURL + "/search"
	body := searchRequest{
		Query:             query,
		SearchDepth:       c.SearchDepth,
		MaxResults:        c.MaxResults,
		IncludeAnswer:     false,
		IncludeRawContent: true,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return port.ProviderResult{Err: fmt.Errorf("tavily: encode request: %w", err)}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return port.ProviderResult{Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if c.ProjectID != "" {
		req.Header.Set("X-Project-ID", c.ProjectID)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return port.ProviderResult{Err: fmt.Errorf("tavily: request: %w", err)}
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return port.ProviderResult{Err: fmt.Errorf("tavily: read body: %w", err)}
	}
	if resp.StatusCode != http.StatusOK {
		return port.ProviderResult{Err: fmt.Errorf("tavily: http %d: %s", resp.StatusCode, truncateForErr(respBody))}
	}

	var out searchResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return port.ProviderResult{Err: fmt.Errorf("tavily: decode response: %w", err)}
	}
	items := make([]port.ProviderItem, 0, len(out.Results))
	for _, r := range out.Results {
		snip := snippetFromTavilyResult(r)
		items = append(items, port.ProviderItem{
			URL:     r.URL,
			Title:   r.Title,
			Snippet: snip,
		})
	}
	return port.ProviderResult{Items: items}
}

func snippetFromTavilyResult(r resultItem) string {
	c := strings.TrimSpace(r.Content)
	raw := strings.TrimSpace(r.RawContent)
	switch {
	case raw != "" && c != "" && raw != c:
		return c + "\n\n" + raw
	case raw != "":
		return raw
	default:
		return c
	}
}

func truncateForErr(b []byte) string {
	const max = 512
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
