package googlebrowser

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

// Client loads Google desktop SERP in headless Chrome and parses #rso / #center_col (see parse.go).
type Client struct {
	MaxResults int
	ExecPath   string
	UserAgent  string
	HL         string // hl= (interface language)
	GL         string // gl= (country/region)
}

// NewFromConfig returns nil unless GOOGLE_BROWSER_ENABLED is true.
func NewFromConfig(cfg config.Config) *Client {
	if !cfg.GoogleBrowserEnabled {
		return nil
	}
	max := cfg.GoogleBrowserMaxResults
	if max < 1 {
		max = 10
	}
	if max > 20 {
		max = 20
	}
	return &Client{
		MaxResults: max,
		ExecPath:   strings.TrimSpace(cfg.ChromeExecPath),
		UserAgent:  strings.TrimSpace(cfg.GoogleBrowserUserAgent),
		HL:         strings.TrimSpace(cfg.GoogleBrowserHL),
		GL:         strings.TrimSpace(cfg.GoogleBrowserGL),
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
	v.Set("num", strconv.Itoa(c.MaxResults))
	if c.HL != "" {
		v.Set("hl", c.HL)
	}
	if c.GL != "" {
		v.Set("gl", c.GL)
	}
	// pws=0: avoid personalized filter in some regions (best-effort).
	v.Set("pws", "0")
	searchURL := "https://www.google.com/search?" + v.Encode()

	var html string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(900*time.Millisecond),
		chromedp.OuterHTML(`#center_col`, &html, chromedp.ByQuery),
	)
	if err != nil {
		return port.ProviderResult{Err: fmt.Errorf("google browser: %w", err)}
	}
	items := parseGoogleResults(html, c.MaxResults)
	if len(items) == 0 {
		return port.ProviderResult{Err: fmt.Errorf("google browser: no organic results parsed (consent/captcha, block, or DOM change)")}
	}
	return port.ProviderResult{Items: items}
}
