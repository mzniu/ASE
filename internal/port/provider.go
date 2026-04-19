package port

import "context"

// ProviderResult is the outcome of a third-party web search API.
type ProviderResult struct {
	Items []ProviderItem
	Err   error
}

// ProviderItem is one search result line (URL + optional snippet).
type ProviderItem struct {
	URL     string
	Title   string
	Snippet string // SERP snippet from the search provider (not full page)
	// BodyMarkdown is optional full-page main content as Markdown (filled after HTTP fetch; see orchestrator).
	BodyMarkdown string
	// Source names which engine produced this row (e.g. baidu, bing); used when multiple providers run.
	Source string
}

// SearchProvider calls external search APIs (Baidu, Bing, Tavily, …).
type SearchProvider interface {
	Search(ctx context.Context, query string) ProviderResult
}
