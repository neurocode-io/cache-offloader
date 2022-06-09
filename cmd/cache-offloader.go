package main

import (
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/client"
	"neurocode.io/cache-offloader/pkg/http"
	"neurocode.io/cache-offloader/pkg/metrics"
	"neurocode.io/cache-offloader/pkg/probes"
	"neurocode.io/cache-offloader/pkg/storage"
)

func getInMemoryStorage(cfg config.Config) http.Cacher {
	switch strings.ToLower(cfg.CacheConfig.Strategy) {
	case "lru":
		return storage.NewLRUCache(cfg.CacheConfig.Size, time.Second*time.Duration(5))
	case "lfu":
		storage.NewLFUCache(cfg.CacheConfig.Size, time.Second*time.Duration(5))
	default:
		log.Fatal().Msgf("Unknown cache strategy: %s. Supported cache strategies are LRU and LFU", cfg.CacheConfig.Strategy)
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
		log.Fatal().Msgf("Unknown cache strategy: %s. Supported cache strategies are LRU and LFU", cfg.CacheConfig.Strategy)
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
	opts := http.ServerOpts{
		Config:           cfg,
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
		log.Fatal().Msgf("Unknown storage: %s. Supported storage options are memory and redis", cfg.ServerConfig.Storage)
	}

	http.RunServer(opts)
}
