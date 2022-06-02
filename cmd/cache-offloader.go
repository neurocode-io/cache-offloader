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
	if strings.ToLower(cfg.CacheConfig.Strategy) == "lru" {
		return storage.NewHashLRU(cfg.MemoryConfig.Size, cfg.CacheConfig)
	} else if strings.ToLower(cfg.CacheConfig.Strategy) == "lfu" {
		// cacher = storage.NewLFUCache(cfg.MemoryConfig.Size)
		// opts.Cacher = storage.NewLFUCache(50)
	} else {
		log.Fatal(fmt.Sprintf("Unknown cache strategy: %s. Supported cache strategies are LRU and LFU", cfg.CacheConfig.Strategy))
	}

	return storage.NewHashLRU(cfg.MemoryConfig.Size, cfg.CacheConfig)
}

func getRedisMemoryStorage(cfg config.Config) http.Cacher {
	r := client.NewRedis(cfg.RedisConfig)
	if strings.ToLower(cfg.CacheConfig.Strategy) == "lru" {
		// configure redis for LRU
	} else if strings.ToLower(cfg.CacheConfig.Strategy) == "lfu" {
		// configure redis for LFU
	} else {
		log.Fatal(fmt.Sprintf("Unknown cache strategy: %s. Supported cache strategies are LRU and LFU", cfg.CacheConfig.Strategy))
	}

	return storage.NewRedisStorage(r.Client, cfg.CacheConfig.CommandTimeoutMilliseconds)
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

	if strings.ToLower(cfg.ServerConfig.Storage) == "memory" {
		opts.Cacher = getInMemoryStorage(cfg)
	} else if strings.ToLower(cfg.ServerConfig.Storage) == "redis" {
		opts.Cacher = getInMemoryStorage(cfg)
		r := client.NewRedis(cfg.RedisConfig)
		opts.ReadinessChecker = storage.NewRedisStorage(r.Client, cfg.CacheConfig.CommandTimeoutMilliseconds)
	} else {
		log.Fatal(fmt.Sprintf("Unknown storage: %s. Supported storage options are memory and redis", cfg.ServerConfig.Storage))
	}

	http.RunServer(opts)
}
