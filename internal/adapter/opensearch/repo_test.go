package opensearch

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ase/internal/config"
)

func TestNewFromConfig_disabled(t *testing.T) {
	r, err := NewFromConfig(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := r.Search(context.Background(), "q")
	if err != nil {
		t.Fatal(err)
	}
	if hits != nil {
		t.Fatalf("noop expected nil hits, got %#v", hits)
	}
}

func TestRepo_Search_parsesHits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/idx/_search" {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(b), "hello") {
			t.Errorf("body %s", b)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "took": 1,
  "hits": {
    "total": { "value": 1, "relation": "eq" },
    "max_score": 2.5,
    "hits": [
      {
        "_index": "idx",
        "_id": "doc-1",
        "_score": 2.5,
        "_source": { "title": "T1", "body_text": "body one two three four five" }
      }
    ]
  }
}`))
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		OpenSearchURLs:       []string{srv.URL},
		OpenSearchIndex:      "idx",
		OpenSearchSearchSize: 5,
		MaxQueryRunes:        100,
	}
	repo, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	r := repo.(*Repo)
	hits, err := r.Search(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("hits %#v", hits)
	}
	if hits[0].ID != "doc-1" || hits[0].Score != 2.5 {
		t.Fatalf("%#v", hits[0])
	}
	if !strings.Contains(hits[0].Body, "T1") || !strings.Contains(hits[0].Body, "body one") {
		t.Fatalf("body %q", hits[0].Body)
	}
}

func TestRepo_IndexDocument_PUTsDoc(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/idx/_doc/d1" {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(b), "my title") || !strings.Contains(string(b), "body text") {
			t.Errorf("body %s", b)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"_index":"idx","_id":"d1","result":"created"}`))
	}))
	t.Cleanup(srv.Close)

	cfg := config.Config{
		OpenSearchURLs:       []string{srv.URL},
		OpenSearchIndex:      "idx",
		OpenSearchSearchSize: 5,
		MaxQueryRunes:        100,
	}
	repo, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	r := repo.(*Repo)
	if err := r.IndexDocument(context.Background(), "d1", "my title", "body text"); err != nil {
		t.Fatal(err)
	}
}

func TestTruncateRunes(t *testing.T) {
	s := strings.Repeat("世", 5)
	got := truncateRunes(s, 3)
	if []rune(got) == nil || len([]rune(got)) != 3 {
		t.Fatalf("got %q", got)
	}
}

func TestComposeHitBody(t *testing.T) {
	if got := composeHitBody("T", "B"); !strings.Contains(got, "T") || !strings.Contains(got, "B") {
		t.Fatal(got)
	}
	if got := composeHitBody("", "only"); got != "only" {
		t.Fatal(got)
	}
}
