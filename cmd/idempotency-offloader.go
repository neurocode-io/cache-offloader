package main

import (
	"fmt"
	h "net/http"
	"net/url"
	"time"

	"dpd.de/idempotency-offloader/config"
	"dpd.de/idempotency-offloader/pkg/client"
	"dpd.de/idempotency-offloader/pkg/http"
	"dpd.de/idempotency-offloader/pkg/storage"
	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
)

func main() {
	config := config.New()
	thisPort := config.ServerConfig.Port
	passthroughEndpoints := config.ServerConfig.PassthroughEndpoints
	downstreamURL, err := url.Parse(config.ServerConfig.DownstreamHost)
	if err != nil {
		log.Panic(fmt.Sprintf("Could not parse downstream url: %s", downstreamURL))
	}
	r := client.NewRedis()
	expirationTime := config.RedisConfig.ExpirationTimeHour * time.Hour
	commandTimeout := config.RedisConfig.CommandTimeoutMillisecond * time.Millisecond
	redisStore := storage.NewRepository(r.Client, &storage.ExpirationTime{Value: expirationTime}, &storage.CommandTimeout{Value: commandTimeout})

	h.HandleFunc("/", http.IdempotencyHandler(redisStore, downstreamURL))
	h.HandleFunc("/probes/readiness", http.ReadinessHandler(redisStore))
	h.HandleFunc("/probes/liveness", http.LivenessHandler)
	h.Handle("/management/prometheus", http.MetricsHandler())

	thisServe := fmt.Sprintf(":%s", thisPort)
	log.Info("Starting idempotency-offloader", rz.String("Port", thisPort))
	log.Info(fmt.Sprintf("Passthrough configured for the following endpoints: %v", passthroughEndpoints))
	log.Panic(fmt.Sprint(h.ListenAndServe(thisServe, nil)))
}
