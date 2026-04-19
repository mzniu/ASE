package noopindex

import (
	"context"

	"github.com/example/ase/internal/port"
)

// Repo returns no hits (OpenSearch not configured or empty index).
type Repo struct{}

func (Repo) Search(_ context.Context, _ string) ([]port.Hit, error) {
	return nil, nil
}

func (Repo) IndexDocument(_ context.Context, _ string, _, _ string) error {
	return port.ErrIndexingDisabled
}
