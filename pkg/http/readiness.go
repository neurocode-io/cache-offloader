package http

import (
	"fmt"
	"net/http"

	"dpd.de/idempotency-offloader/pkg/storage"
	"github.com/bloom42/rz-go/log"
)

func ReadinessHandler(r storage.Repository) http.HandlerFunc {

	return func(res http.ResponseWriter, req *http.Request) {
		err := r.CheckConnection(req.Context())
		if err != nil {
			log.Info("Redis unavailable")
			http.Error(res, "Redis unavailable", http.StatusServiceUnavailable)
			return
		}

		fmt.Fprintf(res, "Alive")

	}
}
