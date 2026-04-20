package domain

import (
	"strings"
	"testing"

	"github.com/example/ase/internal/port"
)

func TestWithoutWritebackIndexHits_filtersPrefix(t *testing.T) {
	in := []port.Hit{
		{ID: "ase-q-deadbeef", Body: "long cached body " + strings.Repeat("x", 100), Score: 9},
		{ID: "manual-1", Body: "short", Score: 1},
	}
	out := WithoutWritebackIndexHits(in, "ase-q-")
	if len(out) != 1 || out[0].ID != "manual-1" {
		t.Fatalf("got %#v", out)
	}
}

func TestWithoutWritebackIndexHits_emptyPrefixUsesDefault(t *testing.T) {
	in := []port.Hit{{ID: "ase-q-abc", Body: "b"}}
	out := WithoutWritebackIndexHits(in, "   ")
	if len(out) != 0 {
		t.Fatalf("want empty, got %#v", out)
	}
}
