package metrics

import (
	"context"
	"errors"

	"github.com/example/ase/internal/orchestrator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// RecordSearchOrchestration counts POST /v1/search outcomes after JSON validation and auth
// (orchestrator errors only; success when err is nil).
func RecordSearchOrchestration(err error) {
	if err == nil {
		searchOutcomes.WithLabelValues("success").Inc()
		return
	}
	outcome := "dependency_unavailable"
	switch {
	case errors.Is(err, orchestrator.ErrBadRequest):
		outcome = "orchestrator_validation"
	case errors.Is(err, context.DeadlineExceeded):
		outcome = "gateway_timeout"
	case errors.Is(err, context.Canceled):
		outcome = "canceled"
	}
	searchOutcomes.WithLabelValues(outcome).Inc()
}

var searchOutcomes = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "ase",
		Name:      "search_orchestration_total",
		Help:      "POST /v1/search orchestration results (after auth and body validation).",
	},
	[]string{"outcome"},
)
