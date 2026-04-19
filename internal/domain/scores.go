package domain

import "github.com/example/ase/internal/port"

// ApplySimilarity sets Hit.Similarity from Hit.Score using batch min–max normalization into [0,1].
func ApplySimilarity(hits []port.Hit) {
	if len(hits) == 0 {
		return
	}
	minS, maxS := hits[0].Score, hits[0].Score
	for i := 1; i < len(hits); i++ {
		s := hits[i].Score
		if s < minS {
			minS = s
		}
		if s > maxS {
			maxS = s
		}
	}
	if maxS == minS {
		for i := range hits {
			hits[i].Similarity = 1
		}
		return
	}
	span := maxS - minS
	for i := range hits {
		hits[i].Similarity = (hits[i].Score - minS) / span
	}
}
