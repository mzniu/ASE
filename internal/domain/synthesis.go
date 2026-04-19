package domain

import "github.com/example/ase/internal/port"

// AgentMarkdownFromIndexHits turns index hits into agent-oriented Markdown (REQ-F-005).
// This is the composition layer: layout and wording for LLM context, not raw SERP HTML parsing.
// Today it delegates to MarkdownFromIndex; stronger strategies (dedupe, cite blocks) can swap in here.
func AgentMarkdownFromIndexHits(query string, hits []port.Hit) string {
	return MarkdownFromIndex(query, hits)
}

// AgentMarkdownFromProviderItems turns provider-normalized items into agent-oriented Markdown (REQ-F-005).
// Preconditions: items come from APIs like Tavily (structured fields), not from scraping an HTML SERP.
// For HTML pages, add an extraction step before building ProviderItem-like structs, then call this.
func AgentMarkdownFromProviderItems(query string, items []port.ProviderItem) string {
	return MarkdownFromProvider(query, items)
}
