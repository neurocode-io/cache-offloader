package main

import (
	"context"
	"fmt"
	"log"
	h "net/http"
	"net/url"
	"time"

	"dpd.de/indempotency-offloader/config"
	"dpd.de/indempotency-offloader/pkg/client"
	"dpd.de/indempotency-offloader/pkg/http"
	"dpd.de/indempotency-offloader/pkg/storage"
)

func main() {

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 100*time.Millisecond)

	config := config.New()
	thisPort := config.ServerConfig.Port
	allowedEndpoints := config.ServerConfig.AllowedEndpoints
	downstreamURL, err := url.Parse(config.ServerConfig.DownstreamHost)
	if err != nil {
		log.Panic(fmt.Sprintf("Could not parse downstream url: %s", downstreamURL))
	}
	r := client.NewRedis()
	redisStore := storage.NewRepository(r.Client)

	h.HandleFunc("/", http.IndempotencyHandler(redisStore, downstreamURL))
	h.HandleFunc("/probes/readiness", http.ReadinessHandler(redisStore))
	h.HandleFunc("/probes/liveness", http.LivenessHandler)

	thisServe := fmt.Sprintf(":%s", thisPort)
	log.Printf("Starting indempotency-offloader, listening: %s", thisServe)
	log.Printf("Indempotency configured for the following endpoints: %v", allowedEndpoints)
	log.Panicf(fmt.Sprint(h.ListenAndServe(thisServe, nil)))
}
