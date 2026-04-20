package admin

import "github.com/example/ase/internal/config"

// ConfigSnapshot returns a JSON-serializable, non-secret view of runtime configuration.
func ConfigSnapshot(cfg config.Config) map[string]any {
	out := map[string]any{
		"http_addr": cfg.HTTPAddr,
		"search": map[string]any{
			"default_providers":   cfg.SearchDefaultProviders,
			"min_hit_count":       cfg.MinHitCount,
			"min_total_text_len":  cfg.MinTotalTextLen,
			"min_similarity":      cfg.MinSimilarity,
			"max_query_runes":     cfg.MaxQueryRunes,
			"max_response_runes":  cfg.MaxResponseRunes,
			"request_deadline_ms": int(cfg.RequestDeadline / 1_000_000),
		},
		"rate_limit": map[string]any{
			"per_key_rps":  cfg.RateLimitPerKey,
			"burst":        cfg.RateLimitBurst,
			"global_rps":   cfg.RateLimitGlobal,
			"global_burst": cfg.RateLimitGlobalBurst,
		},
		"providers": map[string]any{
			"baidu_browser_enabled":  cfg.BaiduBrowserEnabled,
			"bing_browser_enabled":   cfg.BingBrowserEnabled,
			"google_browser_enabled": cfg.GoogleBrowserEnabled,
			"tavily_configured":      cfg.TavilyAPIKey != "",
			"fetch_result_urls":      cfg.ProviderFetchResultURLs,
			"fetch_max_urls":         cfg.ProviderFetchMaxURLs,
		},
		"opensearch": map[string]any{
			"configured":  len(cfg.OpenSearchURLs) > 0 && cfg.OpenSearchIndex != "",
			"index":       cfg.OpenSearchIndex,
			"urls_count":  len(cfg.OpenSearchURLs),
			"search_size": cfg.OpenSearchSearchSize,
		},
		"auth": map[string]any{
			"valid_api_keys_count": len(cfg.AuthValidKeys),
			"dev_api_key_set":      cfg.DevAPIKey != "",
		},
	}
	return out
}
