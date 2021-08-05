package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	thisPort := getEnv("SERVER_PORT", "8000")
	// applicationPort := getEnv("APPLICATION_PORT", "8080")
	allowedEndpoints := getEnvAsSlice("ALLOWED_ENDPOINTS")
	// applicationServer := fmt.Sprintf("http://localhost:%s", applicationPort)

	// downstreamURL, err := url.Parse(applicationServer)
	// if err != nil {
	// 	log.Panic(fmt.Sprintf("Could not parse downstream url: %s", applicationServer))
	// }

	// http.HandleFunc("/", indempotencyHandler(downstreamURL, allowedEndpoints))
	http.HandleFunc("/probes/liveness", livenessHandler)
	http.HandleFunc("/probes/readiness", readinessHandler)

	thisServe := fmt.Sprintf(":%s", thisPort)
	log.Printf("Starting idempotency-offloader, listening: %s", thisServe)
	log.Printf("Idempotency configured for the following endpoints: %v", allowedEndpoints)
	log.Panicf(fmt.Sprint(http.ListenAndServe(thisServe, nil)))
}
