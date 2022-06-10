package config

import (
	"net/url"
	"strings"

	"github.com/rs/zerolog"
)

var logLevel = map[string]zerolog.Level{
	"debug":    zerolog.DebugLevel,
	"info":     zerolog.InfoLevel,
	"warn":     zerolog.WarnLevel,
	"error":    zerolog.ErrorLevel,
	"fatal":    zerolog.FatalLevel,
	"panic":    zerolog.PanicLevel,
	"none":     zerolog.NoLevel,
	"disabled": zerolog.Disabled,
}

type RedisConfig struct {
	ConnectionString string
}

type ServerConfig struct {
	Port        string
	GracePeriod int
	Storage     string // inMemory or redis
	LogLevel    zerolog.Level
}

type CacheConfig struct {
	Strategy        string
	Size            float64
	DownstreamHost  *url.URL
	StaleInSeconds  int
	IgnorePaths     []string
	ShouldHashQuery bool
	HashQueryIgnore map[string]bool
}
type Config struct {
	ServerConfig ServerConfig
	CacheConfig  CacheConfig
	RedisConfig  RedisConfig
}

func New() Config {
	serverConfig := ServerConfig{
		Port:        getEnv("SERVER_PORT", "8000"),
		GracePeriod: getEnvAsInt("SHUTDOWN_GRACE_PERIOD", "30"),
		LogLevel:    getEnvAsLogLevel("SERVER_LOG_LEVEL"),
		Storage:     getEnv("SERVER_STORAGE", ""),
	}

	cacheConfig := CacheConfig{
		Strategy:        getEnv("CACHE_STRATEGY", ""),
		DownstreamHost:  getEnvAsURL("DOWNSTREAM_HOST", ""),
		Size:            getEnvAsFloat("CACHE_SIZE_MB", "10"),
		IgnorePaths:     getEnvAsSlice("CACHE_IGNORE_ENDPOINTS"),
		StaleInSeconds:  getEnvAsInt("CACHE_STALE_WHILE_REVALIDATE_SEC", "5"),
		ShouldHashQuery: getEnvAsBool("CACHE_SHOULD_HASH_QUERY", ""),
		HashQueryIgnore: hashQueryIgnoreMap(getEnvAsSlice("CACHE_HASH_QUERY_IGNORE")),
	}

	if strings.ToLower(serverConfig.Storage) == "memory" {
		return Config{
			ServerConfig: serverConfig,
			CacheConfig:  cacheConfig,
		}
	}

	return Config{
		ServerConfig: serverConfig,
		CacheConfig:  cacheConfig,
		RedisConfig: RedisConfig{
			ConnectionString: getEnv("REDIS_CONNECTION_STRING", ""),
		},
	}
}
