package domain

import (
	"strings"
	"testing"
)

func TestTruncateToRunes_noop(t *testing.T) {
	s := "你好"
	out := TruncateToRunes(s, 10)
	if out != s {
		t.Fatalf("got %q", out)
	}
}

func TestTruncateToRunes_truncates(t *testing.T) {
	s := strings.Repeat("a", 100)
	out := TruncateToRunes(s, 5)
	if !strings.HasSuffix(out, truncatedNotice) {
		t.Fatalf("missing notice: %q", out)
	}
	if len([]rune(out)) <= 5 {
		t.Fatalf("expected longer than 5 runes due to notice")
	}
}
