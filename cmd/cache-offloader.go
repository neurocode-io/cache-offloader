package main

import (
	"time"

	"github.com/skerkour/rz"
	"github.com/skerkour/rz/log"
	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/http"
	"neurocode.io/cache-offloader/pkg/metrics"
	"neurocode.io/cache-offloader/pkg/storage"
)

func main() {
	cfg := config.New()
	log.SetLogger(log.With(rz.Level(cfg.ServerConfig.LogLevel), rz.TimeFieldFormat(time.RFC3339Nano)))

	// r := client.NewRedis(cfg.RedisConfig)

	// commandTimeout := cfg.CacheConfig.CommandTimeoutMilliseconds * time.Millisecond
	// if user wants redistStorage
	// redisCacher := storage.NewRedisStorage(r.Client, commandTimeout)
	// else use inMemoryStorage and algrithm (LRU / LFU)
	inMemoryLRUCacher := storage.NewHashLRU(cfg.MemoryConfig.Size, cfg.CacheConfig)

	// if user wants prometheus metrics
	m := metrics.NewPrometheusCollector()
	// else use noop metrics
	// m := metrics.NewNopMetricsCollector()

	http.RunServer(cfg, inMemoryLRUCacher, m, nil) // inMemoryLRUCacher does not need readinessChecker
}
