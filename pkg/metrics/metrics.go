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

	if err := prometheus.Register(counter); err != nil {
		if alreadyRegistered, ok := err.(prometheus.AlreadyRegisteredError); ok {
			counter = alreadyRegistered.ExistingCollector.(*prometheus.CounterVec)
			return counter
		} else {
			return nil
		}
	}

	return counter
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
