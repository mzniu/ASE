package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"

	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/domain"
	"github.com/example/ase/internal/port"
)

// ErrBadRequest indicates invalid client input (e.g. unknown provider name).
var ErrBadRequest = errors.New("bad request")

// Service wires index-first search and fallback provider (DETAILED_DESIGN §4).
type Service struct {
	Index        port.IndexRepository
	Registry     map[string]port.SearchProvider // keys: baidu, bing, tavily, stub, …
	DefaultNames []string                       // used when the HTTP body omits providers
	Fetcher      port.PageFetcher               // optional; used when fetch is enabled (config and/or deepsearch)
	Config       config.Config

	writeBackSemOnce sync.Once
	writeBackSem     chan struct{} // async index write-back (SEARCH_INDEX_WRITE_BACK_*)
}

// SearchMarkdown returns final Markdown for POST /v1/search.
// providers is optional: nil or empty means DefaultNames; names are case-insensitive (e.g. baidu, bing).
// deepSearch: nil uses Config.ProviderFetchResultURLs; if set, overrides per request (result URL fetch / REQ-F-012).
// indexWrite: nil or true allows async index write-back when Config.SearchIndexWriteBackEnabled; false opts out for this request only.
func (s *Service) SearchMarkdown(ctx context.Context, query string, providers []string, deepSearch *bool, indexWrite *bool) (string, error) {
	rid := middleware.GetReqID(ctx)
	if rid == "" {
		rid = "-"
	}
	qRunes := utf8.RuneCountInString(query)
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, s.Config.RequestDeadline)
	defer cancel()

	hits, err := s.Index.Search(ctx, query)
	if err != nil {
		slog.Error("index search failed", "request_id", rid, "query_runes", qRunes, "err", err)
		return "", fmt.Errorf("index search: %w", err)
	}
	wbPrefix := strings.TrimSpace(s.Config.SearchIndexWriteBackIDPrefix)
	if wbPrefix == "" {
		wbPrefix = domain.DefaultWritebackHitIDPrefix
	}
	hits = domain.WithoutWritebackIndexHits(hits, wbPrefix)
	domain.ApplySimilarity(hits)
	if domain.Enough(hits, s.Config.MinHitCount, s.Config.MinTotalTextLen, s.Config.MinSimilarity) {
		md := domain.AgentMarkdownFromIndexHits(query, hits)
		md = domain.TruncateToRunes(md, s.Config.MaxResponseRunes)
		slog.Info("search ok",
			"request_id", rid,
			"path", "index",
			"hits", len(hits),
			"query_runes", qRunes,
			"out_runes", utf8.RuneCountInString(md),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return md, nil
	}

	names, err := s.resolveProviderNames(providers)
	if err != nil {
		return "", err
	}

	items, provErr := s.runProvidersParallel(ctx, query, names)
	if provErr != nil {
		slog.Error("provider search failed", "request_id", rid, "query_runes", qRunes, "err", provErr)
		return "", fmt.Errorf("provider search: %w", provErr)
	}
	items = domain.MergeProviderItemsDedupe(items)

	fetchURLs := s.Config.ProviderFetchResultURLs
	if deepSearch != nil {
		fetchURLs = *deepSearch
	}
	var pages []port.FetchedPage
	if fetchURLs && s.Fetcher != nil && len(items) > 0 {
		urls := make([]string, 0, len(items))
		for _, it := range items {
			if it.URL != "" {
				urls = append(urls, it.URL)
			}
		}
		pages = s.Fetcher.FetchPlainText(ctx, urls, s.Config.ProviderFetchMaxURLs)
		items = domain.EnrichProviderItemsWithFetch(items, pages)
	}
	nFetchPages := len(pages)
	nFetchWithText := 0
	for _, p := range pages {
		if strings.TrimSpace(p.Text) != "" {
			nFetchWithText++
		}
	}
	nBody := 0
	for _, it := range items {
		if strings.TrimSpace(it.BodyMarkdown) != "" {
			nBody++
		}
	}

	md := domain.AgentMarkdownFromProviderItems(query, items)
	md = domain.TruncateToRunes(md, s.Config.MaxResponseRunes)
	slog.Info("search ok",
		"request_id", rid,
		"path", "provider",
		"providers", strings.Join(names, ","),
		"provider_results", len(items),
		"query_runes", qRunes,
		"fetch_pages", nFetchPages,
		"fetch_nonempty", nFetchWithText,
		"body_markdown_items", nBody,
		"out_runes", utf8.RuneCountInString(md),
		"duration_ms", time.Since(start).Milliseconds(),
	)
	allowWB := s.Config.SearchIndexWriteBackEnabled && (indexWrite == nil || *indexWrite)
	s.scheduleProviderIndexWriteBack(query, md, rid, allowWB)
	return md, nil
}

func (s *Service) resolveProviderNames(requested []string) ([]string, error) {
	names := requested
	if len(names) == 0 {
		names = s.DefaultNames
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no providers: %w", ErrBadRequest)
	}
	var out []string
	seen := make(map[string]struct{})
	for _, raw := range names {
		n := strings.ToLower(strings.TrimSpace(raw))
		if n == "" {
			continue
		}
		if _, ok := s.Registry[n]; !ok {
			return nil, fmt.Errorf("unknown provider %q: %w", n, ErrBadRequest)
		}
		if _, dup := seen[n]; dup {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid providers in request: %w", ErrBadRequest)
	}
	return out, nil
}

func (s *Service) runProvidersParallel(ctx context.Context, query string, names []string) ([]port.ProviderItem, error) {
	g, ctx := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	byName := make(map[string][]port.ProviderItem, len(names))
	var fail []error

	for _, name := range names {
		name := name
		p := s.Registry[name]
		g.Go(func() error {
			pr := p.Search(ctx, query)
			if pr.Err != nil {
				mu.Lock()
				fail = append(fail, fmt.Errorf("%s: %w", name, pr.Err))
				mu.Unlock()
				slog.Warn("provider returned error", "provider", name, "err", pr.Err)
				return nil
			}
			mu.Lock()
			items := make([]port.ProviderItem, 0, len(pr.Items))
			for _, it := range pr.Items {
				it.Source = name
				items = append(items, it)
			}
			byName[name] = items
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	var merged []port.ProviderItem
	for _, name := range names {
		merged = append(merged, byName[name]...)
	}
	if len(merged) == 0 && len(fail) > 0 {
		return nil, errors.Join(fail...)
	}
	return merged, nil
}
