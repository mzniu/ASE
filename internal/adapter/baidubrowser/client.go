package baidubrowser

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

// Client loads Baidu desktop SERP in headless Chrome and parses #content_left (see parse.go).
type Client struct {
	MaxResults int
	ExecPath   string // optional path to chrome/chromium (CHROME_EXEC_PATH)
	UserAgent  string
}

// NewFromConfig returns nil unless BAIDU_BROWSER_ENABLED is true.
func NewFromConfig(cfg config.Config) *Client {
	if !cfg.BaiduBrowserEnabled {
		return nil
	}
	max := cfg.BaiduBrowserMaxResults
	if max < 1 {
		max = 10
	}
	if max > 20 {
		max = 20
	}
	return &Client{
		MaxResults: max,
		ExecPath:   strings.TrimSpace(cfg.ChromeExecPath),
		UserAgent:  strings.TrimSpace(cfg.BaiduBrowserUserAgent),
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
		chromedp.Flag("lang", "zh-CN"),
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
	v.Set("wd", q)
	v.Set("rn", strconv.Itoa(c.MaxResults))
	searchURL := "https://www.baidu.com/s?" + v.Encode()

	var html string
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`#content_left`, chromedp.ByQuery),
		chromedp.Sleep(550*time.Millisecond),
		chromedp.OuterHTML(`#content_left`, &html, chromedp.ByQuery),
	)
	if err != nil {
		return port.ProviderResult{Err: fmt.Errorf("baidu browser: %w", err)}
	}
	items := parseBaiduContentLeft(html, c.MaxResults)
	if len(items) == 0 {
		return port.ProviderResult{Err: fmt.Errorf("baidu browser: no organic results parsed (captcha, block, or DOM change)")}
	}
	return port.ProviderResult{Items: items}
}
