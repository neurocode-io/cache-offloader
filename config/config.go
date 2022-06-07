package config

import (
	"fmt"
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
	Password         string
	Database         int
	Size             int
}

type MemoryConfig struct {
	Size float64
}

type ServerConfig struct {
	Port           string
	GracePeriod    int
	DownstreamHost string
	Storage        string // inMemory or redis
	LogLevel       zerolog.Level
}

type CacheConfig struct {
	Strategy        string
	StaleInSeconds  int
	IgnorePaths     []string
	ShouldHashQuery bool
	HashQueryIgnore map[string]bool
}
type Config struct {
	ServerConfig ServerConfig
	CacheConfig  CacheConfig
	RedisConfig  RedisConfig
	MemoryConfig MemoryConfig
}

func New() Config {
	serverConfig := ServerConfig{
		Port:           getEnv("SERVER_PORT", "8000"),
		GracePeriod:    getEnvAsInt("SHUTDOWN_GRACE_PERIOD", "30"),
		DownstreamHost: getEnv("DOWNSTREAM_HOST", ""),
		LogLevel:       getEnvAsLogLevel("SERVER_LOG_LEVEL"),
		Storage:        getEnv("SERVER_STORAGE", ""),
	}

	cacheConfig := CacheConfig{
		Strategy:        getEnv("CACHE_STRATEGY", ""),
		IgnorePaths:     getEnvAsSlice("CACHE_IGNORE_ENDPOINTS"),
		StaleInSeconds:  getEnvAsInt("CACHE_STALE_WHILE_REVALIDATE_SEC", "5"),
		ShouldHashQuery: getEnvAsBool("CACHE_SHOULD_HASH_QUERY", ""),
		HashQueryIgnore: hashQueryIgnoreMap(getEnvAsSlice("CACHE_HASH_QUERY_IGNORE")),
	}

	if strings.ToLower(serverConfig.Storage) == "memory" {
		return Config{
			ServerConfig: serverConfig,
			CacheConfig:  cacheConfig,
			MemoryConfig: MemoryConfig{
				Size: getEnvAsFloat("MEMORY_CACHE_SIZE_MB", "50"),
			},
		}
	}

	return Config{
		ServerConfig: serverConfig,
		CacheConfig:  cacheConfig,
		RedisConfig: RedisConfig{
			ConnectionString: fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", ""), getEnv("REDIS_PORT", "")),
			Password:         getEnv("REDIS_PASSWORD", ""),
			Database:         getEnvAsInt("REDIS_DB", "0"),
			Size:             getEnvAsInt("REDIS_CACHE_SIZE_MB", "10"),
		},
	}
}
