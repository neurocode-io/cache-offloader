package main

import (
	"context"
	"fmt"
	"log"
	h "net/http"
	"net/url"
	"time"

	"dpd.de/idempotency-offloader/config"
	"dpd.de/idempotency-offloader/pkg/client"
	"dpd.de/idempotency-offloader/pkg/http"
	"dpd.de/idempotency-offloader/pkg/storage"
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
	expirationTime := config.RedisConfig.ExpirationTimeHour * time.Hour
	commandTimeout := config.RedisConfig.CommandTimeoutMillisecond * time.Millisecond
	redisStore := storage.NewRepository(r.Client, commandTimeout, expirationTime)

	h.HandleFunc("/", http.IdempotencyHandler(redisStore, downstreamURL))
	h.HandleFunc("/probes/readiness", http.ReadinessHandler(redisStore))
	h.HandleFunc("/probes/liveness", http.LivenessHandler)

	thisServe := fmt.Sprintf(":%s", thisPort)
	log.Printf("Starting idempotency-offloader, listening: %s", thisServe)
	log.Printf("Idempotency configured for the following endpoints: %v", allowedEndpoints)
	log.Panicf(fmt.Sprint(h.ListenAndServe(thisServe, nil)))
}
