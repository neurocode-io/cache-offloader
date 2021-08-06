package main

import (
	"fmt"
	"log"
	h "net/http"

	"dpd.de/indempotency-offloader/config"
	"dpd.de/indempotency-offloader/pkg/http"
)

func main() {

	thisPort := config.GetEnv("SERVER_PORT", "8000")
	// applicationPort := getEnv("APPLICATION_PORT", "8080")
	allowedEndpoints := config.GetEnvAsSlice("ALLOWED_ENDPOINTS")
	// applicationServer := fmt.Sprintf("http://localhost:%s", applicationPort)

	// downstreamURL, err := url.Parse(applicationServer)
	// if err != nil {
	// 	log.Panic(fmt.Sprintf("Could not parse downstream url: %s", applicationServer))
	// }

	// http.HandleFunc("/", indempotencyHandler(downstreamURL, allowedEndpoints))

	h.HandleFunc("/probes/liveness", http.LivenessHandler)
	h.HandleFunc("/probes/readiness", http.ReadinessHandler)

	thisServe := fmt.Sprintf(":%s", thisPort)
	log.Printf("Starting idempotency-offloader, listening: %s", thisServe)
	log.Printf("Idempotency configured for the following endpoints: %v", allowedEndpoints)
	log.Panicf(fmt.Sprint(h.ListenAndServe(thisServe, nil)))
}
