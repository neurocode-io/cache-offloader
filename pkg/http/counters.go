package http

import "dpd.de/idempotency-offloader/pkg/metrics"

var (
	StatusCounter = metrics.AddCounterVec(
		&metrics.CounterVecOpts{
			Name:       "http_request_to_downstream_urls",
			Help:       "Number of downstream requests by status.",
			LabelNames: "status",
		},
	)

	ResponseSourceCounter = metrics.AddCounterVec(
		&metrics.CounterVecOpts{
			Name:       "http_responses",
			Help:       "Number of http responses with status<500 by source.",
			LabelNames: "source",
		},
	)
)
