package domain

import (
	"testing"

	"github.com/example/ase/internal/port"
)

func TestEnough_falseLowCount(t *testing.T) {
	h := []port.Hit{{Body: "hello world", Similarity: 1}}
	if Enough(h, 2, 5, 0) {
		t.Fatal("expected false")
	}
}

func TestEnough_falseLowLen(t *testing.T) {
	h := []port.Hit{{Body: "hi", Similarity: 1}}
	if Enough(h, 1, 10, 0) {
		t.Fatal("expected false")
	}
}

func TestEnough_falseLowSimilarity(t *testing.T) {
	h := []port.Hit{{Body: "hello world", Similarity: 0.1}}
	if Enough(h, 1, 5, 0.5) {
		t.Fatal("expected false")
	}
}

func TestEnough_true(t *testing.T) {
	h := []port.Hit{{Body: "hello world wide", Similarity: 0.9}}
	if !Enough(h, 1, 10, 0.5) {
		t.Fatal("expected true")
	}
}
