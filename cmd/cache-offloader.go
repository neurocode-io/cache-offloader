package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/skerkour/rz"
	"github.com/skerkour/rz/log"
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
		return storage.NewHashLRU(cfg.MemoryConfig.Size)
	case "lfu":
		// cacher = storage.NewLFUCache(cfg.MemoryConfig.Size)
		// opts.Cacher = storage.NewLFUCache(50)
	default:
		log.Fatal(fmt.Sprintf("Unknown cache strategy: %s. Supported cache strategies are LRU and LFU", cfg.CacheConfig.Strategy))
	}

	return nil
}

func getRedisMemoryStorage(cfg config.Config) http.Cacher {
	r := client.NewRedis(cfg.RedisConfig)
	switch strings.ToLower(cfg.CacheConfig.Strategy) {
	case "lru":
		// configure redis for LRU cache
		return storage.NewRedisStorage(r.Client)
	case "lfu":
		// configure redis for LFU cache
		return storage.NewRedisStorage(r.Client)
	default:
		log.Fatal(fmt.Sprintf("Unknown cache strategy: %s. Supported cache strategies are LRU and LFU", cfg.CacheConfig.Strategy))
	}

	return nil
}

func main() {
	cfg := config.New()
	log.SetLogger(log.With(rz.Level(cfg.ServerConfig.LogLevel), rz.TimeFieldFormat(time.RFC3339Nano)))
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
		opts.Cacher = getRedisMemoryStorage(cfg)
		r := client.NewRedis(cfg.RedisConfig)
		opts.ReadinessChecker = storage.NewRedisStorage(r.Client)
	default:
		log.Fatal(fmt.Sprintf("Unknown storage: %s. Supported storage options are memory and redis", cfg.ServerConfig.Storage))
	}

	http.RunServer(opts)
}
