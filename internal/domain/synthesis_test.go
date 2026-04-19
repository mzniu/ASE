package domain

import (
	"testing"

	"github.com/example/ase/internal/port"
)

func TestAgentMarkdownFromProviderItems_delegates(t *testing.T) {
	items := []port.ProviderItem{{Snippet: "x"}}
	a := AgentMarkdownFromProviderItems("q", items)
	b := MarkdownFromProvider("q", items)
	if a != b {
		t.Fatal("expected same output as MarkdownFromProvider in MVP")
	}
}

func TestAgentMarkdownFromIndexHits_delegates(t *testing.T) {
	h := []port.Hit{{Body: "body"}}
	a := AgentMarkdownFromIndexHits("q", h)
	b := MarkdownFromIndex("q", h)
	if a != b {
		t.Fatal("expected same output as MarkdownFromIndex in MVP")
	}
}
