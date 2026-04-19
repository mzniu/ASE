package domain

import (
	"strings"
	"testing"

	"github.com/example/ase/internal/port"
)

func TestMarkdownFromProvider_multilineSnippet(t *testing.T) {
	md := MarkdownFromProvider("q", []port.ProviderItem{{
		Snippet:      "first para",
		BodyMarkdown: "second block\nline",
	}})
	if !strings.Contains(md, "## 摘要") || !strings.Contains(md, "## 正文") {
		t.Fatal(md)
	}
	if !strings.Contains(md, "first para") || !strings.Contains(md, "second block") {
		t.Fatal(md)
	}
}

func TestMarkdownFromProvider_includesURL(t *testing.T) {
	md := MarkdownFromProvider("q", []port.ProviderItem{{
		Title:   "T",
		URL:     "https://example.com/p",
		Snippet: "body line",
	}})
	if !strings.Contains(md, "**链接**：https://example.com/p") || !strings.Contains(md, "body line") {
		t.Fatal(md)
	}
}
