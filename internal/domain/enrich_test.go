package domain

import (
	"strings"
	"testing"

	"github.com/example/ase/internal/port"
)

func TestEnrichProviderItemsWithFetch_appendsExcerpt(t *testing.T) {
	items := []port.ProviderItem{
		{URL: "https://a.example/x", Title: "t", Snippet: "api line"},
	}
	pages := []port.FetchedPage{
		{URL: "https://a.example/x", Text: "long body from page"},
	}
	out := EnrichProviderItemsWithFetch(items, pages)
	if len(out) != 1 {
		t.Fatal(len(out))
	}
	if out[0].Snippet != "api line" {
		t.Fatalf("snippet = %q", out[0].Snippet)
	}
	if !strings.Contains(out[0].BodyMarkdown, "long body") {
		t.Fatalf("BodyMarkdown = %q", out[0].BodyMarkdown)
	}
}

func TestEnrichProviderItemsWithFetch_urlKeyMatchesHTTPVariant(t *testing.T) {
	items := []port.ProviderItem{
		{URL: "HTTP://a.example/x", Snippet: "s"},
	}
	pages := []port.FetchedPage{
		{URL: "http://a.example/x", Text: "fetched"},
	}
	out := EnrichProviderItemsWithFetch(items, pages)
	if out[0].BodyMarkdown != "fetched" {
		t.Fatalf("BodyMarkdown = %q", out[0].BodyMarkdown)
	}
}

func TestEnrichProviderItemsWithFetch_noPages(t *testing.T) {
	items := []port.ProviderItem{{URL: "u", Snippet: "s"}}
	out := EnrichProviderItemsWithFetch(items, nil)
	if len(out) != 1 || out[0].Snippet != "s" {
		t.Fatalf("%#v", out)
	}
}
