package orchestrator

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// indexWriteBackTotal counts async index write-back attempts after provider-path /v1/search.
// Labels: skipped_global_off, skipped_request_optout, skipped_body_short, ok, error, noop.
var indexWriteBackTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "ase",
		Name:      "index_writeback_total",
		Help:      "OpenSearch async index write-back outcomes (Phase-1/2; provider path only).",
	},
	[]string{"result"},
)
