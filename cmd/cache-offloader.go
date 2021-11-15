package main

import (
	"fmt"
	h "net/http"
	"net/url"
	"time"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/client"
	"neurocode.io/cache-offloader/pkg/http"
	"neurocode.io/cache-offloader/pkg/storage"
)

func main() {
	config := config.New()
	log.SetLogger(log.With(rz.Level(config.ServerConfig.LogLevel), rz.TimeFieldFormat(time.RFC3339Nano)))
	thisPort := config.ServerConfig.Port
	passthroughEndpoints := config.ServerConfig.PassthroughEndpoints
	downstreamURL, err := url.Parse(config.ServerConfig.DownstreamHost)
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not parse downstream url: %s", downstreamURL))
	}

	r := client.NewRedis()
	expirationTime := config.RedisConfig.ExpirationTimeMinutes * time.Minute
	commandTimeout := config.RedisConfig.CommandTimeoutMillisecond * time.Millisecond
	redisStore := storage.NewRepository(r.Client, &storage.ExpirationTime{Value: expirationTime}, &storage.CommandTimeout{Value: commandTimeout})

	h.HandleFunc("/", http.CacheHandler(redisStore, downstreamURL))
	h.HandleFunc("/probes/readiness", http.ReadinessHandler(redisStore))
	h.HandleFunc("/probes/liveness", http.LivenessHandler)
	h.Handle("/management/prometheus", http.MetricsHandler())

	thisServe := fmt.Sprintf(":%s", thisPort)
	log.Info(fmt.Sprintf("Starting cache-offloader on port %v", thisPort))
	log.Info(fmt.Sprintf("Downstream host configured %v", downstreamURL))

	log.Info(fmt.Sprintf("Passthrough endpoints configured: %v", passthroughEndpoints), rz.Timestamp(true))

	err = h.ListenAndServe(thisServe, nil)
	log.Fatal("Unhandled error occured. Will reboot.", rz.Err(err))
}
