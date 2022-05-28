package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/skerkour/rz/log"
)

type ReadinessChecker interface {
	CheckConnection(context.Context) error
}

func readinessHandler(r ReadinessChecker) http.HandlerFunc {

	return func(res http.ResponseWriter, req *http.Request) {
		err := r.CheckConnection(req.Context())
		if err != nil {
			log.Warn("Redis unavailable")
			http.Error(res, "Redis unavailable", http.StatusServiceUnavailable)
			return
		}

		fmt.Fprintf(res, "Alive")

	}
}
