package domain

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWritebackBodyFromMarkdown_stripsLinksAndFences(t *testing.T) {
	in := "# Title\n\nSee [here](https://x.com) and **bold**.\n\n```\ncode\n```\n\nTail."
	out := WritebackBodyFromMarkdown(in, 500)
	if strings.Contains(out, "https://x.com") {
		t.Fatalf("URL should be stripped: %q", out)
	}
	if strings.Contains(out, "```") {
		t.Fatalf("fence should be removed: %q", out)
	}
	if !strings.Contains(out, "here") || !strings.Contains(out, "bold") {
		t.Fatalf("expected kept words: %q", out)
	}
}

func TestWritebackBodyFromMarkdown_truncate(t *testing.T) {
	in := strings.Repeat("字", 100)
	out := WritebackBodyFromMarkdown(in, 10)
	if utf8.RuneCountInString(out) != 10 {
		t.Fatalf("len = %d", utf8.RuneCountInString(out))
	}
}
