package orchestrator

import (
	"context"
	"errors"
	"strings"
	"sync"
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

type captureIndex struct {
	sliceIndex
	mu     sync.Mutex
	calls  int
	lastID string
	wg     sync.WaitGroup
}

func (c *captureIndex) IndexDocument(_ context.Context, id, _, _ string) error {
	c.mu.Lock()
	c.calls++
	c.lastID = id
	c.mu.Unlock()
	c.wg.Done()
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
	md, err := svc.SearchMarkdown(context.Background(), "q", nil, nil, nil)
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
	_, err := svc.SearchMarkdown(context.Background(), "q", nil, nil, nil)
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
	md, err := svc.SearchMarkdown(context.Background(), "q", nil, nil, nil)
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

func TestService_deepSearch_overridesProviderFetch(t *testing.T) {
	idx := sliceIndex{hits: nil}
	sp := &spyProvider{result: port.ProviderResult{Items: []port.ProviderItem{
		{URL: "https://example.com/a", Snippet: "snippet"},
	}}}
	fetcher := fixedFetcher{pages: []port.FetchedPage{{URL: "https://example.com/a", Text: "deep body"}}}
	base := func(fetchCfg bool) *Service {
		return &Service{
			Index:        idx,
			Registry:     map[string]port.SearchProvider{"stub": sp},
			DefaultNames: []string{"stub"},
			Fetcher:      fetcher,
			Config: config.Config{
				MinHitCount:             1,
				MinTotalTextLen:         10,
				ProviderFetchResultURLs: fetchCfg,
				ProviderFetchMaxURLs:    2,
				MaxResponseRunes:        10000,
				RequestDeadline:         time.Minute,
			},
		}
	}

	off := false
	md, err := base(true).SearchMarkdown(context.Background(), "q", nil, &off, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(md, "deep body") {
		t.Fatal("expected no fetch when deepsearch=false")
	}

	on := true
	md, err = base(false).SearchMarkdown(context.Background(), "q", nil, &on, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "deep body") {
		t.Fatalf("expected fetch when deepsearch=true: %q", md)
	}
}

func TestService_indexError(t *testing.T) {
	idx := sliceIndex{err: errors.New("down")}
	svc := &Service{
		Index:  idx,
		Config: config.Config{MaxResponseRunes: 100, RequestDeadline: time.Minute},
	}
	_, err := svc.SearchMarkdown(context.Background(), "q", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestService_providerPath_indexWriteBack_queryKeyedID(t *testing.T) {
	idx := &captureIndex{sliceIndex: sliceIndex{hits: nil}}
	idx.wg.Add(1)
	snippet := strings.Repeat("snippet text ", 8)
	sp := &spyProvider{result: port.ProviderResult{Items: []port.ProviderItem{{Snippet: snippet}}}}
	svc := &Service{
		Index:        idx,
		Registry:     map[string]port.SearchProvider{"stub": sp},
		DefaultNames: []string{"stub"},
		Config: config.Config{
			MinHitCount:                        1,
			MinTotalTextLen:                    1000,
			MinSimilarity:                      0,
			MaxResponseRunes:                   10000,
			RequestDeadline:                    time.Minute,
			SearchIndexWriteBackEnabled:        true,
			SearchIndexWriteBackTimeout:        5 * time.Second,
			SearchIndexWriteBackMinBodyRunes:   10,
			SearchIndexWriteBackMaxBodyRunes:   100000,
			SearchIndexWriteBackTitleMaxRunes:  200,
			SearchIndexWriteBackIDPrefix:       "ase-q-",
			SearchIndexWriteBackMaxConcurrency: 4,
		},
	}
	q := "hello-writeback-query-unique"
	_, err := svc.SearchMarkdown(context.Background(), q, []string{"stub"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		idx.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for IndexDocument")
	}
	want := queryWriteBackDocID("ase-q-", q)
	idx.mu.Lock()
	got := idx.lastID
	idx.mu.Unlock()
	if got != want {
		t.Fatalf("doc id = %q want %q", got, want)
	}
}

func TestService_providerPath_indexWriteBack_requestOptOut(t *testing.T) {
	idx := &captureIndex{sliceIndex: sliceIndex{hits: nil}}
	snippet := strings.Repeat("snippet text ", 8)
	sp := &spyProvider{result: port.ProviderResult{Items: []port.ProviderItem{{Snippet: snippet}}}}
	svc := &Service{
		Index:        idx,
		Registry:     map[string]port.SearchProvider{"stub": sp},
		DefaultNames: []string{"stub"},
		Config: config.Config{
			MinHitCount:                        1,
			MinTotalTextLen:                    1000,
			MinSimilarity:                      0,
			MaxResponseRunes:                   10000,
			RequestDeadline:                    time.Minute,
			SearchIndexWriteBackEnabled:        true,
			SearchIndexWriteBackTimeout:        5 * time.Second,
			SearchIndexWriteBackMinBodyRunes:   10,
			SearchIndexWriteBackMaxBodyRunes:   100000,
			SearchIndexWriteBackTitleMaxRunes:  200,
			SearchIndexWriteBackIDPrefix:       "ase-q-",
			SearchIndexWriteBackMaxConcurrency: 4,
		},
	}
	optOut := false
	_, err := svc.SearchMarkdown(context.Background(), "opt-out-query", []string{"stub"}, nil, &optOut)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	idx.mu.Lock()
	n := idx.calls
	idx.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected no IndexDocument when index_write=false, calls=%d", n)
	}
}

func TestService_indexEnough_skipsWriteBack(t *testing.T) {
	idx := &captureIndex{sliceIndex: sliceIndex{hits: []port.Hit{{Body: strings.Repeat("a", 200), Score: 1}}}}
	svc := &Service{
		Index:        idx,
		Registry:     map[string]port.SearchProvider{"stub": &spyProvider{}},
		DefaultNames: []string{"stub"},
		Config: config.Config{
			MinHitCount:                 1,
			MinTotalTextLen:             10,
			MinSimilarity:               0,
			MaxResponseRunes:            10000,
			RequestDeadline:             time.Minute,
			SearchIndexWriteBackEnabled: true,
		},
	}
	_, err := svc.SearchMarkdown(context.Background(), "q", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(150 * time.Millisecond)
	idx.mu.Lock()
	n := idx.calls
	idx.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected no IndexDocument on index path, calls=%d", n)
	}
}
