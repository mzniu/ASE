package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds service configuration loaded from the environment.
type Config struct {
	HTTPAddr string

	// AuthValidKeys lists accepted Bearer tokens (comma-separated). If empty and DevAPIKey is set, only DevAPIKey matches.
	AuthValidKeys []string
	DevAPIKey     string

	MinHitCount      int
	MinTotalTextLen  int     // UTF-8 runes summed over hit bodies
	MinSimilarity    float64 // on normalized scores
	MaxResponseRunes int
	MaxQueryRunes    int
	RequestDeadline  time.Duration

	RateLimitPerKey      float64 // sustained QPS per key
	RateLimitBurst       int
	RateLimitGlobal      float64
	RateLimitGlobalBurst int

	// Tavily (optional fallback SearchProvider)
	TavilyAPIKey      string
	TavilyBaseURL     string // e.g. https://api.tavily.com
	TavilyMaxResults  int
	TavilySearchDepth string // basic | advanced | fast | ultra-fast
	TavilyProjectID   string // optional X-Project-ID

	// Optional: after SearchProvider returns, fetch https result URLs and append plain-text excerpts (see AGENT_MARKDOWN_PIPELINE.md).
	ProviderFetchResultURLs bool
	ProviderFetchMaxURLs    int
	FetchPerURLTimeoutMs    int
	FetchConcurrency        int // max parallel HTTPS GETs when fetching result pages

	// OpenSearch (self-hosted index; DETAILED_DESIGN §6.3). Empty URLs or index name disables real client (noop index).
	OpenSearchURLs       []string
	OpenSearchIndex      string
	OpenSearchUser       string
	OpenSearchPassword   string
	OpenSearchSearchSize int // _search size parameter

	// Baidu SERP via headless Chrome (chromedp). Takes precedence over Tavily when enabled.
	BaiduBrowserEnabled    bool
	BaiduBrowserMaxResults int
	ChromeExecPath         string // optional; default: find chrome on PATH / well-known paths
	BaiduBrowserUserAgent  string

	// Bing SERP via headless Chrome (chromedp), same pattern as Baidu.
	BingBrowserEnabled    bool
	BingBrowserMaxResults int
	BingBrowserUserAgent  string
	BingBrowserMarket     string // e.g. zh-CN → mkt= (optional)

	// Google SERP via headless Chrome (chromedp).
	GoogleBrowserEnabled    bool
	GoogleBrowserMaxResults int
	GoogleBrowserUserAgent  string
	GoogleBrowserHL         string // hl= (e.g. zh-CN)
	GoogleBrowserGL         string // gl= (e.g. cn)

	// SearchDefaultProviders is the default list when the JSON body omits "providers" (comma env SEARCH_DEFAULT_PROVIDERS).
	SearchDefaultProviders []string

	// Admin UI (GET /admin/): username + password (bcrypt hash or dev plain) + session secret.
	AdminUsername       string
	AdminPasswordBcrypt string // env ADMIN_PASSWORD_BCRYPT (preferred)
	AdminPasswordPlain  string // env ADMIN_PASSWORD — dev only; ignored when AdminPasswordBcrypt is set
	AdminSessionSecret  string // env ADMIN_SESSION_SECRET — HMAC key for cookie session
	AdminSessionTTL     time.Duration
}

// Load reads configuration from environment variables with documented defaults.
func Load() Config {
	cfg := Config{
		HTTPAddr:                getenv("HTTP_ADDR", ":18080"),
		DevAPIKey:               os.Getenv("DEV_API_KEY"),
		MinHitCount:             getenvInt("MIN_HIT_COUNT", 1),
		MinTotalTextLen:         getenvInt("MIN_TOTAL_TEXT_LEN", 20),
		MinSimilarity:           getenvFloat("MIN_SIMILARITY", 0),
		MaxResponseRunes:        getenvInt("MAX_RESPONSE_RUNES", 16000),
		MaxQueryRunes:           getenvInt("MAX_QUERY_RUNES", 4096),
		RequestDeadline:         getenvDuration("REQUEST_DEADLINE_MS", 55*time.Second),
		RateLimitPerKey:         getenvFloat("RATE_LIMIT_PER_KEY_RPS", 10),
		RateLimitBurst:          getenvInt("RATE_LIMIT_BURST", 20),
		RateLimitGlobal:         getenvFloat("RATE_LIMIT_GLOBAL_RPS", 100),
		RateLimitGlobalBurst:    getenvInt("RATE_LIMIT_GLOBAL_BURST", 200),
		TavilyAPIKey:            os.Getenv("TAVILY_API_KEY"),
		TavilyBaseURL:           getenv("TAVILY_BASE_URL", "https://api.tavily.com"),
		TavilyMaxResults:        getenvInt("TAVILY_MAX_RESULTS", 10),
		TavilySearchDepth:       getenv("TAVILY_SEARCH_DEPTH", "basic"),
		TavilyProjectID:         os.Getenv("TAVILY_PROJECT_ID"),
		ProviderFetchResultURLs: getenvBool("PROVIDER_FETCH_RESULT_URLS", false),
		ProviderFetchMaxURLs:    getenvInt("PROVIDER_FETCH_MAX_URLS", 2),
		FetchPerURLTimeoutMs:    getenvInt("FETCH_PER_URL_TIMEOUT_MS", 8000),
		FetchConcurrency:        getenvInt("FETCH_CONCURRENCY", 4),
		OpenSearchIndex:         strings.TrimSpace(os.Getenv("OPENSEARCH_INDEX")),
		OpenSearchUser:          os.Getenv("OPENSEARCH_USER"),
		OpenSearchPassword:      os.Getenv("OPENSEARCH_PASSWORD"),
		OpenSearchSearchSize:    getenvInt("OPENSEARCH_SEARCH_SIZE", 10),
		BaiduBrowserEnabled:     getenvBool("BAIDU_BROWSER_ENABLED", false),
		BaiduBrowserMaxResults:  getenvInt("BAIDU_BROWSER_MAX_RESULTS", 10),
		ChromeExecPath:          os.Getenv("CHROME_EXEC_PATH"),
		BaiduBrowserUserAgent:   os.Getenv("BAIDU_BROWSER_USER_AGENT"),
		BingBrowserEnabled:      getenvBool("BING_BROWSER_ENABLED", false),
		BingBrowserMaxResults:   getenvInt("BING_BROWSER_MAX_RESULTS", 10),
		BingBrowserUserAgent:    os.Getenv("BING_BROWSER_USER_AGENT"),
		BingBrowserMarket:       getenv("BING_BROWSER_MARKET", "zh-CN"),
		GoogleBrowserEnabled:    getenvBool("GOOGLE_BROWSER_ENABLED", false),
		GoogleBrowserMaxResults: getenvInt("GOOGLE_BROWSER_MAX_RESULTS", 10),
		GoogleBrowserUserAgent:  os.Getenv("GOOGLE_BROWSER_USER_AGENT"),
		GoogleBrowserHL:         getenv("GOOGLE_BROWSER_HL", ""),
		GoogleBrowserGL:         getenv("GOOGLE_BROWSER_GL", ""),
		AdminUsername:           strings.TrimSpace(os.Getenv("ADMIN_USERNAME")),
		AdminPasswordBcrypt:     strings.TrimSpace(os.Getenv("ADMIN_PASSWORD_BCRYPT")),
		AdminPasswordPlain:      os.Getenv("ADMIN_PASSWORD"),
		AdminSessionSecret:      strings.TrimSpace(os.Getenv("ADMIN_SESSION_SECRET")),
	}
	ttlSec := getenvInt("ADMIN_SESSION_TTL_SECONDS", 86400)
	if ttlSec < 60 {
		ttlSec = 86400
	}
	cfg.AdminSessionTTL = time.Duration(ttlSec) * time.Second
	if s := os.Getenv("SEARCH_DEFAULT_PROVIDERS"); s != "" {
		for _, p := range strings.Split(s, ",") {
			p = strings.TrimSpace(strings.ToLower(p))
			if p != "" {
				cfg.SearchDefaultProviders = append(cfg.SearchDefaultProviders, p)
			}
		}
	}
	if s := os.Getenv("OPENSEARCH_URLS"); s != "" {
		for _, p := range strings.Split(s, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.OpenSearchURLs = append(cfg.OpenSearchURLs, p)
			}
		}
	}
	if s := os.Getenv("AUTH_VALID_API_KEYS"); s != "" {
		for _, p := range strings.Split(s, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.AuthValidKeys = append(cfg.AuthValidKeys, p)
			}
		}
	}
	return cfg
}

// AdminUIEnabled is true when admin routes should be registered (username, password, and session secret present).
func (c Config) AdminUIEnabled() bool {
	if c.AdminUsername == "" || c.AdminSessionSecret == "" {
		return false
	}
	if len(c.AdminSessionSecret) < 16 {
		return false
	}
	if c.AdminPasswordBcrypt != "" {
		return true
	}
	return strings.TrimSpace(c.AdminPasswordPlain) != ""
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func getenvBool(key string, def bool) bool {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	switch strings.ToLower(s) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func getenvFloat(key string, def float64) float64 {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return f
}

func getenvDuration(key string, def time.Duration) time.Duration {
	ms := getenvInt(key, 0)
	if ms <= 0 {
		return def
	}
	return time.Duration(ms) * time.Millisecond
}
