// server.go
package http

import (
	"context"
	"fmt"
	h "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/neurocode-io/cache-offloader/config"
)

type ServerOpts struct {
	Config           config.Config
	Cacher           Cacher
	Worker           Worker
	MetricsCollector MetricsCollector
	ReadinessChecker ReadinessChecker
}

func RunServer(opts ServerOpts) {
	mux := h.NewServeMux()

	// Cache handler at root
	mux.Handle("/", newCacheHandler(opts.Cacher, opts.MetricsCollector, opts.Worker, opts.Config.CacheConfig))
	// Local endpoints
	mux.Handle("/metrics/prometheus", metricsHandler())
	mux.HandleFunc("/probes/liveness", livenessHandler)
	mux.HandleFunc("/probes/readiness", readinessHandler(opts.ReadinessChecker))

	// Forward ignored paths directly to downstream (no caching)
	fwd := newForwardHandler(opts.Config.CacheConfig.DownstreamHost)
	for _, path := range opts.Config.CacheConfig.IgnorePaths {
		if path == "/metrics/prometheus" || path == "/probes/liveness" || path == "/probes/readiness" {
			continue
		}
		mux.Handle(path, fwd)
	}

	server := &h.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%s", opts.Config.ServerConfig.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,   // protect against slowloris
		ReadTimeout:       0,                 // allow streaming requests
		WriteTimeout:      0,                 // don’t cut off long responses
		IdleTimeout:       120 * time.Second, // keep-alive; match LB/ingress
		MaxHeaderBytes:    8 << 20,           // optional: 8MB
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	log.Info().Msgf("Starting server on port %s", opts.Config.ServerConfig.Port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		log.Warn().Msg("received interrupt signal, shutting down...")

		shutdownCtx, cancel := context.WithTimeout(serverCtx, time.Duration(opts.Config.ServerConfig.GracePeriod)*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal().Msg("graceful shutdown timed out.. forcing exit.")
				cancel()
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal().Stack().Err(err).Msg("error shutting down server")
		}
		serverStopCtx()
	}()

	if err := server.ListenAndServe(); err != nil && err != h.ErrServerClosed {
		log.Fatal().Stack().Err(err).Msg("error running server")
	}
	<-serverCtx.Done()
}
