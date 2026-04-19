package bingbrowser

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/port"
)

// Client loads Bing desktop SERP in headless Chrome and parses #b_results (see parse.go).
type Client struct {
	MaxResults int
	ExecPath   string
	UserAgent  string
	Market     string // e.g. zh-CN for mkt=
}

// NewFromConfig returns nil unless BING_BROWSER_ENABLED is true.
func NewFromConfig(cfg config.Config) *Client {
	if !cfg.BingBrowserEnabled {
		return nil
	}
	max := cfg.BingBrowserMaxResults
	if max < 1 {
		max = 10
	}
	if max > 20 {
		max = 20
	}
	return &Client{
		MaxResults: max,
		ExecPath:   strings.TrimSpace(cfg.ChromeExecPath),
		UserAgent:  strings.TrimSpace(cfg.BingBrowserUserAgent),
		Market:     strings.TrimSpace(cfg.BingBrowserMarket),
	}
}

// Search implements port.SearchProvider.
func (c *Client) Search(ctx context.Context, query string) port.ProviderResult {
	q := strings.TrimSpace(query)
	if q == "" {
		return port.ProviderResult{Items: nil}
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	if c.ExecPath != "" {
		opts = append(opts, chromedp.ExecPath(c.ExecPath))
	}
	ua := c.UserAgent
	if ua == "" {
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 ASE/1.0"
	}
	opts = append(opts, chromedp.UserAgent(ua))

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	v := url.Values{}
	v.Set("q", q)
	v.Set("count", strconv.Itoa(c.MaxResults))
	if m := c.Market; m != "" {
		v.Set("mkt", m)
	}
	searchURL := "https://www.bing.com/search?" + v.Encode()

	var html string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`#b_results`, chromedp.ByQuery),
		chromedp.Sleep(550*time.Millisecond),
		chromedp.OuterHTML(`#b_results`, &html, chromedp.ByQuery),
	)
	if err != nil {
		return port.ProviderResult{Err: fmt.Errorf("bing browser: %w", err)}
	}
	items := parseBingResults(html, c.MaxResults)
	if len(items) == 0 {
		return port.ProviderResult{Err: fmt.Errorf("bing browser: no organic results parsed (block, consent wall, or DOM change)")}
	}
	return port.ProviderResult{Items: items}
}
