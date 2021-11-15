package http

import (
	"fmt"
	"net/http"

	"github.com/bloom42/rz-go/log"
	"neurocode.io/cache-offloader/pkg/storage"
)

func ReadinessHandler(r storage.Repository) http.HandlerFunc {

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
