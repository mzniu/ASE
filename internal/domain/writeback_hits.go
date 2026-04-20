package domain

import (
	"strings"

	"github.com/example/ase/internal/port"
)

// DefaultWritebackHitIDPrefix is the default OpenSearch _id prefix for async query write-back documents.
const DefaultWritebackHitIDPrefix = "ase-q-"

// WithoutWritebackIndexHits drops hits whose _id uses the query write-back scheme (SEARCH_INDEX_WRITE_BACK_ID_PREFIX + SHA256(query)).
// Those docs live in the same index as corpus hits; with MIN_SIMILARITY=0 a single large write-back doc can otherwise satisfy Enough
// for unrelated queries (often the only hit, ApplySimilarity → similarity 1). Excluding them restores provider fallback for freshness.
func WithoutWritebackIndexHits(hits []port.Hit, writebackIDPrefix string) []port.Hit {
	p := strings.TrimSpace(writebackIDPrefix)
	if p == "" {
		p = DefaultWritebackHitIDPrefix
	}
	out := make([]port.Hit, 0, len(hits))
	for _, h := range hits {
		if strings.HasPrefix(h.ID, p) {
			continue
		}
		out = append(out, h)
	}
	return out
}
