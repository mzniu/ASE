package port

import "context"

// FetchedPage is best-effort main content for one URL (HTTPS GET).
type FetchedPage struct {
	URL  string
	Text string // article body as Markdown when using readability + HTML→Markdown; empty if skipped or failed
}

// PageFetcher performs controlled HTTP fetches for result URLs (REQ-F-008, DETAILED_DESIGN §5.3).
type PageFetcher interface {
	// FetchPlainText fetches up to limit URLs in order (first distinct http/https URLs in the slice).
	// Implementations must respect ctx cancellation/deadline and avoid unbounded concurrency.
	FetchPlainText(ctx context.Context, urls []string, limit int) []FetchedPage
}
