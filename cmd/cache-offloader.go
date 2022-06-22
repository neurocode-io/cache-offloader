package main

import (
	"strings"
	"time"

	"github.com/neurocode-io/cache-offloader/config"
	"github.com/neurocode-io/cache-offloader/pkg/client"
	"github.com/neurocode-io/cache-offloader/pkg/http"
	"github.com/neurocode-io/cache-offloader/pkg/metrics"
	"github.com/neurocode-io/cache-offloader/pkg/probes"
	"github.com/neurocode-io/cache-offloader/pkg/storage"
	"github.com/neurocode-io/cache-offloader/pkg/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func getInMemoryStorage(cfg config.Config) http.Cacher {
	switch strings.ToLower(cfg.CacheConfig.Strategy) {
	case "lru":
		return storage.NewLRUCache(cfg.CacheConfig.Size, cfg.CacheConfig.StaleInSeconds)
	case "lfu":
		return storage.NewLFUCache(cfg.CacheConfig.Size, cfg.CacheConfig.StaleInSeconds)
	default:
		log.Error().Msgf("Unknown cache strategy: %s", cfg.CacheConfig.Strategy)
		log.Fatal().Msg("Supported cache strategies are LRU and LFU")
	}

	return nil
}

func getRedisMemoryStorage(cfg config.Config) storage.RedisStorage {
	r := client.NewRedis(cfg.RedisConfig)
	switch strings.ToLower(cfg.CacheConfig.Strategy) {
	case "lru":
		r.ConfigureLRU(cfg.CacheConfig.Size)

		return storage.NewRedisStorage(r.Client, cfg.CacheConfig.StaleInSeconds)
	case "lfu":
		r.ConfigureLFU(cfg.CacheConfig.Size)

		return storage.NewRedisStorage(r.Client, cfg.CacheConfig.StaleInSeconds)
	default:
		log.Error().Msgf("Unknown cache strategy: %s", cfg.CacheConfig.Strategy)
		log.Fatal().Msgf("Supported cache strategies are LRU and LFU")
	}

	return storage.RedisStorage{}
}

func setupLogging(logLevel zerolog.Level) {
	zerolog.SetGlobalLevel(logLevel)
	zerolog.MessageFieldName = "msg"
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	l := log.Level(logLevel)
	zerolog.DefaultContextLogger = &l
}

func main() {
	cfg := config.New()
	setupLogging(cfg.ServerConfig.LogLevel)
	m := metrics.NewPrometheusCollector()
	maxInFlightRevalidationRequests := 1000
	opts := http.ServerOpts{
		Config:           cfg,
		Worker:           worker.NewUpdateQueue(maxInFlightRevalidationRequests),
		MetricsCollector: m,
		ReadinessChecker: probes.NewReadinessChecker(),
	}

	switch strings.ToLower(cfg.ServerConfig.Storage) {
	case "memory":
		opts.Cacher = getInMemoryStorage(cfg)
	case "redis":
		cacher := getRedisMemoryStorage(cfg)
		opts.Cacher = cacher
		opts.ReadinessChecker = cacher
	default:
		log.Error().Msgf("Unknown storage: %s", cfg.ServerConfig.Storage)
		log.Fatal().Msg("Supported storages are memory and redis")
	}

	http.RunServer(opts)
}
