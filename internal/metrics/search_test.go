package metrics

import (
	"context"
	"errors"
	"testing"

	"github.com/example/ase/internal/orchestrator"
)

func TestRecordSearchOrchestration_coversOutcomes(t *testing.T) {
	t.Parallel()
	RecordSearchOrchestration(nil)
	RecordSearchOrchestration(orchestrator.ErrBadRequest)
	RecordSearchOrchestration(context.DeadlineExceeded)
	RecordSearchOrchestration(context.Canceled)
	RecordSearchOrchestration(errors.New("generic upstream"))
}
