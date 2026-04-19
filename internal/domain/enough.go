package domain

import (
	"unicode/utf8"

	"github.com/example/ase/internal/port"
)

// Enough reports whether index hits satisfy thresholds (REQ-F-007).
func Enough(hits []port.Hit, minCount, minTotalRunes int, minSimilarity float64) bool {
	if len(hits) < minCount {
		return false
	}
	total := 0
	maxSim := 0.0
	for _, h := range hits {
		total += utf8.RuneCountInString(h.Body)
		if h.Similarity > maxSim {
			maxSim = h.Similarity
		}
	}
	if total < minTotalRunes {
		return false
	}
	if maxSim < minSimilarity {
		return false
	}
	return true
}
