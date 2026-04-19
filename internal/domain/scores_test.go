package domain

import (
	"math"
	"testing"

	"github.com/example/ase/internal/port"
)

func TestApplySimilarity_singleHit(t *testing.T) {
	h := []port.Hit{{Score: 3.5, Body: "x"}}
	ApplySimilarity(h)
	if h[0].Similarity != 1 {
		t.Fatalf("Similarity = %v, want 1", h[0].Similarity)
	}
}

func TestApplySimilarity_minMax(t *testing.T) {
	h := []port.Hit{
		{Score: 1, Body: "a"},
		{Score: 3, Body: "b"},
	}
	ApplySimilarity(h)
	if math.Abs(h[0].Similarity-0) > 1e-9 {
		t.Fatalf("first Similarity = %v", h[0].Similarity)
	}
	if math.Abs(h[1].Similarity-1) > 1e-9 {
		t.Fatalf("second Similarity = %v", h[1].Similarity)
	}
}

func TestApplySimilarity_equalScores(t *testing.T) {
	h := []port.Hit{{Score: 2, Body: "a"}, {Score: 2, Body: "b"}}
	ApplySimilarity(h)
	if h[0].Similarity != 1 || h[1].Similarity != 1 {
		t.Fatalf("got %+v %+v", h[0].Similarity, h[1].Similarity)
	}
}
