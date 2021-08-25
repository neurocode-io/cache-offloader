package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CounterVecOpts struct {
	Name       string
	Help       string
	LabelNames string
}

func AddCounterVec(opts *CounterVecOpts) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: opts.Name,
			Help: opts.Help,
		},
		[]string{opts.LabelNames},
	)

	prometheus.MustRegister(counter)

	return counter
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
