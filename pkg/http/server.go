package http

import (
	"context"
	"fmt"
	h "net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"neurocode.io/cache-offloader/config"
)

type ServerOpts struct {
	Config           config.Config
	Cacher           Cacher
	MetricsCollector MetricsCollector
	ReadinessChecker ReadinessChecker
}

func RunServer(opts ServerOpts) {
	downstreamURL, err := url.Parse(opts.Config.ServerConfig.DownstreamHost)
	if err != nil {
		log.Fatal().Msgf("Could not parse downstream url: %s", downstreamURL)
	}

	mux := h.NewServeMux()
	mux.Handle("/", newCacheHandler(opts.Cacher, opts.MetricsCollector, *downstreamURL))
	mux.Handle("/metrics/prometheus", metricsHandler())
	mux.HandleFunc("/probes/liveness", livenessHandler)

	for _, path := range opts.Config.CacheConfig.IgnorePaths {
		mux.HandleFunc(path, forwardHandler(downstreamURL))
	}

	mux.HandleFunc("/probes/readiness", readinessHandler(opts.ReadinessChecker))

	server := &h.Server{Addr: fmt.Sprintf("0.0.0.0:%s", opts.Config.ServerConfig.Port), Handler: mux}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	log.Info().Msgf("Starting server on port %s", opts.Config.ServerConfig.Port)

	// Listen for syscall signals for process to interrupt/quit
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

		// Trigger graceful shutdown
		err = server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("error shutting down server")
		}
		serverStopCtx()
	}()

	// Run the server
	err = server.ListenAndServe()
	if err != nil && err != h.ErrServerClosed {
		log.Fatal().Stack().Err(err).Msg("error running server")
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
