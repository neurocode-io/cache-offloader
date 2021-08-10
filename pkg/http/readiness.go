package http

import (
	"fmt"
	"log"
	"net/http"

	"dpd.de/indempotency-offloader/pkg/storage"
)

func ReadinessHandler(r storage.Repository) http.HandlerFunc {

	return func(res http.ResponseWriter, req *http.Request) {
		err := r.CheckConnection(req.Context())
		if err != nil {
			log.Println("Redis unavailable")
			http.Error(res, "Redis unavailable", 503)
			return
		}

		fmt.Fprintf(res, "Alive")

	}
}
