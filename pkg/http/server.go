package http

import (
	"context"
	"fmt"
	h "net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/skerkour/rz"
	"github.com/skerkour/rz/log"
	"neurocode.io/cache-offloader/config"
)

func RunServer(cfg config.Config, cacher Cacher, metricsCollector MetricsCollector, redinessChecker ReadinessChecker) {
	downstreamURL, err := url.Parse(cfg.ServerConfig.DownstreamHost)
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not parse downstream url: %s", downstreamURL))
	}

	mux := h.NewServeMux()
	mux.Handle("/", newStaleWhileRevalidateHandler(cacher, metricsCollector, *downstreamURL))
	mux.Handle("/metrics/prometheus", metricsHandler())
	mux.HandleFunc("/probes/liveness", livenessHandler)

	for _, path := range cfg.CacheConfig.IgnorePaths {
		log.Info(path)
		mux.HandleFunc(path, forwardHandler(downstreamURL))
	}

	if strings.ToLower(cfg.ServerConfig.Storage) == "redis" {
		mux.HandleFunc("/probes/readiness", readinessHandler(redinessChecker))
	}

	server := &h.Server{Addr: fmt.Sprintf("0.0.0.0:%s", cfg.ServerConfig.Port), Handler: mux}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	log.Info(fmt.Sprintf("Starting server on port %s", cfg.ServerConfig.Port))

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		log.Warn("received interrupt signal, shutting down...")

		shutdownCtx, cancel := context.WithTimeout(serverCtx, time.Duration(cfg.ServerConfig.GracePeriod)*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
				cancel()
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal("", rz.Stack(true), rz.Err(err))
		}
		serverStopCtx()
	}()

	// Run the server
	err = server.ListenAndServe()
	if err != nil && err != h.ErrServerClosed {
		log.Fatal("", rz.Stack(true), rz.Err(err))
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
