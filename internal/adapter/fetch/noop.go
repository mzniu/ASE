package fetch

import (
	"context"

	"github.com/example/ase/internal/port"
)

// Noop implements port.PageFetcher without performing any network I/O.
type Noop struct{}

// FetchPlainText returns nil (no enrichment).
func (Noop) FetchPlainText(_ context.Context, _ []string, _ int) []port.FetchedPage {
	return nil
}
