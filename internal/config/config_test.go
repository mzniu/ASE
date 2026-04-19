package config

import (
	"testing"
	"time"
)

func TestLoad_defaults(t *testing.T) {
	for _, k := range []string{
		"HTTP_ADDR", "AUTH_VALID_API_KEYS", "DEV_API_KEY",
		"MIN_HIT_COUNT", "MIN_TOTAL_TEXT_LEN", "MIN_SIMILARITY",
		"MAX_RESPONSE_RUNES", "MAX_QUERY_RUNES", "REQUEST_DEADLINE_MS",
		"RATE_LIMIT_PER_KEY_RPS", "RATE_LIMIT_BURST",
		"RATE_LIMIT_GLOBAL_RPS", "RATE_LIMIT_GLOBAL_BURST",
		"TAVILY_API_KEY", "TAVILY_BASE_URL", "TAVILY_MAX_RESULTS", "TAVILY_SEARCH_DEPTH", "TAVILY_PROJECT_ID",
		"PROVIDER_FETCH_RESULT_URLS", "PROVIDER_FETCH_MAX_URLS", "FETCH_PER_URL_TIMEOUT_MS", "FETCH_CONCURRENCY",
		"OPENSEARCH_URLS", "OPENSEARCH_INDEX", "OPENSEARCH_USER", "OPENSEARCH_PASSWORD", "OPENSEARCH_SEARCH_SIZE",
		"BAIDU_BROWSER_ENABLED", "BAIDU_BROWSER_MAX_RESULTS", "CHROME_EXEC_PATH", "BAIDU_BROWSER_USER_AGENT",
		"BING_BROWSER_ENABLED", "BING_BROWSER_MAX_RESULTS", "BING_BROWSER_USER_AGENT", "BING_BROWSER_MARKET",
		"GOOGLE_BROWSER_ENABLED", "GOOGLE_BROWSER_MAX_RESULTS", "GOOGLE_BROWSER_USER_AGENT", "GOOGLE_BROWSER_HL", "GOOGLE_BROWSER_GL",
		"SEARCH_DEFAULT_PROVIDERS",
	} {
		t.Setenv(k, "")
	}
	c := Load()
	if c.HTTPAddr != ":18080" {
		t.Fatalf("HTTPAddr = %q", c.HTTPAddr)
	}
	if c.MinHitCount != 1 {
		t.Fatalf("MinHitCount = %d", c.MinHitCount)
	}
	if c.RequestDeadline != 55*time.Second {
		t.Fatalf("RequestDeadline = %v", c.RequestDeadline)
	}
	if c.FetchConcurrency != 4 {
		t.Fatalf("FetchConcurrency = %d", c.FetchConcurrency)
	}
}

func TestLoad_authKeys(t *testing.T) {
	t.Setenv("AUTH_VALID_API_KEYS", "a, b , ")
	t.Setenv("DEV_API_KEY", "")
	c := Load()
	if len(c.AuthValidKeys) != 2 || c.AuthValidKeys[0] != "a" || c.AuthValidKeys[1] != "b" {
		t.Fatalf("keys = %#v", c.AuthValidKeys)
	}
}

func TestLoad_openSearchURLs(t *testing.T) {
	t.Setenv("OPENSEARCH_URLS", " https://a:9200 , https://b:9200 ")
	t.Setenv("OPENSEARCH_INDEX", "idx1")
	c := Load()
	if len(c.OpenSearchURLs) != 2 || c.OpenSearchURLs[0] != "https://a:9200" {
		t.Fatalf("OpenSearchURLs = %#v", c.OpenSearchURLs)
	}
	if c.OpenSearchIndex != "idx1" {
		t.Fatal(c.OpenSearchIndex)
	}
}
