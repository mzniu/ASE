package domain

import (
	"strings"
	"testing"

	"github.com/example/ase/internal/port"
)

func TestMergeProviderItemsDedupe_sameURL(t *testing.T) {
	items := []port.ProviderItem{
		{URL: "https://example.com/a", Title: "T", Snippet: "one", Source: "baidu"},
		{URL: "HTTPS://example.com/a", Title: "T2", Snippet: "two", Source: "bing"},
	}
	out := MergeProviderItemsDedupe(items)
	if len(out) != 1 {
		t.Fatalf("len=%d", len(out))
	}
	if !strings.Contains(out[0].Snippet, "one") || !strings.Contains(out[0].Snippet, "two") {
		t.Fatalf("snippet=%q", out[0].Snippet)
	}
	if !strings.Contains(out[0].Source, "baidu") || !strings.Contains(out[0].Source, "bing") {
		t.Fatalf("source=%q", out[0].Source)
	}
}
