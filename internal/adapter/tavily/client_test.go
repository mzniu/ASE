package tavily

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ase/internal/config"
)

func TestNewFromConfig_nilWithoutKey(t *testing.T) {
	if NewFromConfig(config.Config{}) != nil {
		t.Fatal("expected nil")
	}
}

func TestClient_Search_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "auth", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"A","url":"https://a.example","content":"snippet text","raw_content":"raw more"}]}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		HTTP:        srv.Client(),
		APIKey:      "tvly-test",
		BaseURL:     srv.URL,
		MaxResults:  5,
		SearchDepth: "basic",
	}
	res := c.Search(context.Background(), "hello")
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("items = %d", len(res.Items))
	}
	it := res.Items[0]
	if it.URL != "https://a.example" || !strings.Contains(it.Snippet, "snippet text") || !strings.Contains(it.Snippet, "raw more") {
		t.Fatalf("item = %+v", it)
	}
}

func TestClient_Search_httpError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "quota", http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		HTTP:        srv.Client(),
		APIKey:      "k",
		BaseURL:     srv.URL,
		MaxResults:  3,
		SearchDepth: "basic",
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
		TavilyAPIKey:     "x",
		TavilyMaxResults: 99,
	})
	if c.MaxResults != 20 {
		t.Fatalf("MaxResults = %d", c.MaxResults)
	}
}
