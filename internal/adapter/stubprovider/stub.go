package stubprovider

import (
	"context"

	"github.com/example/ase/internal/port"
)

// Fixed returns a constant ProviderResult (for bootstrap / tests).
type Fixed struct {
	Result port.ProviderResult
}

func (f Fixed) Search(_ context.Context, _ string) port.ProviderResult {
	return f.Result
}
