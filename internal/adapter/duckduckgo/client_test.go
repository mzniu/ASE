package duckduckgo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/ase/internal/config"
)

func TestNewFromConfig_nilWhenDisabled(t *testing.T) {
	if NewFromConfig(config.Config{DuckDuckGoEnabled: false}) != nil {
		t.Fatal("expected nil")
	}
}

func TestClient_Search_parsesHTMLFixture(t *testing.T) {
	html := `<!DOCTYPE html><html><body><div class="results"><div class="result results_links">
  <h2 class="result__title"><a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2F">Example Co</a></h2>
  <a class="result__snippet" href="https://example.com/">First line of snippet.</a>
</div><div class="result results_links">
  <h2 class="result__title"><a class="result__a" href="https://other.test/page">Other</a></h2>
</div></div></body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/html" {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.Form.Get("q") == "" {
			http.Error(w, "missing q", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		HTTP:       srv.Client(),
		BaseURL:    srv.URL + "/html",
		MaxResults: 10,
		UserAgent:  "ASE-test/1.0",
	}
	res := c.Search(context.Background(), "hello world")
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if len(res.Items) != 2 {
		t.Fatalf("items = %d %+v", len(res.Items), res.Items)
	}
	if res.Items[0].URL != "https://example.com/" || res.Items[0].Title != "Example Co" {
		t.Fatalf("first = %+v", res.Items[0])
	}
	if !strings.Contains(res.Items[0].Snippet, "First line") {
		t.Fatalf("snippet = %q", res.Items[0].Snippet)
	}
	if res.Items[1].URL != "https://other.test/page" {
		t.Fatalf("second = %+v", res.Items[1])
	}
}

func TestClient_Search_httpError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "slow down", http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)
	c := &Client{
		HTTP:       srv.Client(),
		BaseURL:    srv.URL,
		MaxResults: 3,
		UserAgent:  "x",
	}
	res := c.Search(context.Background(), "q")
	if res.Err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(res.Err.Error(), "429") {
		t.Fatalf("err = %v", res.Err)
	}
}

func TestNewFromConfig_clampsMaxResults(t *testing.T) {
	c := NewFromConfig(config.Config{
		DuckDuckGoEnabled:    true,
		DuckDuckGoMaxResults: 999,
	})
	if c.MaxResults != 25 {
		t.Fatalf("MaxResults = %d", c.MaxResults)
	}
}

func TestNewFromConfig_usesTimeout(t *testing.T) {
	c := NewFromConfig(config.Config{
		DuckDuckGoEnabled:   true,
		DuckDuckGoTimeoutMs: 5000,
	})
	if c.HTTP.Timeout != 5*time.Second {
		t.Fatalf("timeout = %v", c.HTTP.Timeout)
	}
}
