package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/ase/internal/config"
	"github.com/example/ase/internal/port"
)

type sliceIndex struct {
	hits []port.Hit
	err  error
}

func (s sliceIndex) Search(_ context.Context, _ string) ([]port.Hit, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.hits, nil
}

func (sliceIndex) IndexDocument(_ context.Context, _ string, _, _ string) error {
	return nil
}

type fixedFetcher struct {
	pages []port.FetchedPage
}

func (f fixedFetcher) FetchPlainText(_ context.Context, _ []string, limit int) []port.FetchedPage {
	if limit <= 0 || len(f.pages) == 0 {
		return nil
	}
	if len(f.pages) > limit {
		return f.pages[:limit]
	}
	return f.pages
}

type spyProvider struct {
	calls  int
	result port.ProviderResult
}

func (s *spyProvider) Search(_ context.Context, _ string) port.ProviderResult {
	s.calls++
	return s.result
}

func TestService_indexEnough_skipsProvider(t *testing.T) {
	idx := sliceIndex{hits: []port.Hit{
		{Body: "hello world wide enough text here", Score: 5},
	}}
	svc := &Service{
		Index: idx,
		Config: config.Config{
			MinHitCount:      1,
			MinTotalTextLen:  10,
			MinSimilarity:    0,
			MaxResponseRunes: 10000,
			RequestDeadline:  time.Minute,
		},
	}
	md, err := svc.SearchMarkdown(context.Background(), "q", nil)
	if err != nil {
		t.Fatal(err)
	}
	if md == "" {
		t.Fatal("empty markdown")
	}
}

func TestService_fallback_callsProvider(t *testing.T) {
	idx := sliceIndex{hits: nil}
	sp := &spyProvider{result: port.ProviderResult{Items: []port.ProviderItem{{Snippet: "from api"}}}}
	svc := &Service{
		Index: idx,
		Registry: map[string]port.SearchProvider{
			"stub": sp,
		},
		DefaultNames: []string{"stub"},
		Config: config.Config{
			MinHitCount:      1,
			MinTotalTextLen:  10,
			MinSimilarity:    0,
			MaxResponseRunes: 10000,
			RequestDeadline:  time.Minute,
		},
	}
	_, err := svc.SearchMarkdown(context.Background(), "q", nil)
	if err != nil {
		t.Fatal(err)
	}
	if sp.calls != 1 {
		t.Fatalf("provider calls = %d", sp.calls)
	}
}

func TestService_providerFetch_enrichesMarkdown(t *testing.T) {
	idx := sliceIndex{hits: nil}
	sp := &spyProvider{result: port.ProviderResult{Items: []port.ProviderItem{
		{URL: "https://example.com/a", Snippet: "snippet from api"},
	}}}
	svc := &Service{
		Index: idx,
		Registry: map[string]port.SearchProvider{
			"stub": sp,
		},
		DefaultNames: []string{"stub"},
		Fetcher: fixedFetcher{pages: []port.FetchedPage{
			{URL: "https://example.com/a", Text: "fetched plain text body"},
		}},
		Config: config.Config{
			MinHitCount:             1,
			MinTotalTextLen:         10,
			ProviderFetchResultURLs: true,
			ProviderFetchMaxURLs:    2,
			MaxResponseRunes:        10000,
			RequestDeadline:         time.Minute,
		},
	}
	md, err := svc.SearchMarkdown(context.Background(), "q", nil)
	if err != nil {
		t.Fatal(err)
	}
	if sp.calls != 1 {
		t.Fatalf("provider calls = %d", sp.calls)
	}
	for _, sub := range []string{"snippet from api", "fetched plain text body", "## 正文"} {
		if !strings.Contains(md, sub) {
			t.Fatalf("markdown missing %q: %q", sub, md)
		}
	}
}

func TestService_indexError(t *testing.T) {
	idx := sliceIndex{err: errors.New("down")}
	svc := &Service{
		Index:  idx,
		Config: config.Config{MaxResponseRunes: 100, RequestDeadline: time.Minute},
	}
	_, err := svc.SearchMarkdown(context.Background(), "q", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
