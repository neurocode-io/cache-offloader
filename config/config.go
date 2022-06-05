package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/skerkour/rz"
)

var logLevel = map[string]rz.LogLevel{
	"debug":    rz.DebugLevel,
	"info":     rz.InfoLevel,
	"warn":     rz.WarnLevel,
	"error":    rz.ErrorLevel,
	"fatal":    rz.FatalLevel,
	"panic":    rz.PanicLevel,
	"none":     rz.NoLevel,
	"disabled": rz.Disabled,
}

type RedisConfig struct {
	ConnectionString string
	Password         string
	Database         int
	Size             int
	Algorithm        string
}

type MemoryConfig struct {
	Size      float64
	Algorithm string
}

type ServerConfig struct {
	Port           string
	GracePeriod    int
	DownstreamHost string
	Storage        string // inMemory or redis
	LogLevel       rz.LogLevel
}

type CacheConfig struct {
	Strategy             string
	StaleWhileRevalidate int
	CommandTimeout       time.Duration
	IgnorePaths          []string
	HashShouldQuery      bool
	HashQueryIgnore      map[string]bool
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
		GracePeriod:    getEnvAsInt("CACHE_STALE_WHILE_REVALIDATE_SEC", "30"),
		DownstreamHost: getEnv("DOWNSTREAM_HOST", ""),
		LogLevel:       getEnvAsLogLevel("SERVER_LOG_LEVEL"),
		Storage:        getEnv("SERVER_STORAGE", ""),
	}

	if strings.ToLower(serverConfig.Storage) == "memory" {
		return Config{
			ServerConfig: serverConfig,
			CacheConfig: CacheConfig{
				Strategy:             getEnv("CACHE_STRATEGY", ""),
				IgnorePaths:          getEnvAsSlice("CACHE_IGNORE_ENDPOINTS"),
				StaleWhileRevalidate: getEnvAsInt("CACHE_STALE_WHILE_REVALIDATE_SEC", "5"),
				HashShouldQuery:      getEnvAsBool("CACHE_SHOULD_HASH_QUERY", ""),
				HashQueryIgnore:      hashQueryIgnoreMap(getEnvAsSlice("CACHE_HASH_QUERY_IGNORE")),
			},
			MemoryConfig: MemoryConfig{
				Size:      getEnvAsFloat("MEMORY_CACHE_SIZE_MB", "50"),
				Algorithm: strings.ToLower(getEnv("MEMORY_CACHE_ALGORITHM", "LRU")),
			},
		}
	}

	return Config{
		ServerConfig: serverConfig,
		CacheConfig: CacheConfig{
			Strategy:             getEnv("CACHE_STRATEGY", ""),
			IgnorePaths:          getEnvAsSlice("CACHE_IGNORE_ENDPOINTS"),
			StaleWhileRevalidate: getEnvAsInt("CACHE_STALE_WHILE_REVALIDATE_SEC", "5"),
			HashShouldQuery:      getEnvAsBool("CACHE_SHOULD_HASH_QUERY", ""),
			HashQueryIgnore:      hashQueryIgnoreMap(getEnvAsSlice("CACHE_HASH_QUERY_IGNORE")),
			CommandTimeout:       time.Duration(getEnvAsInt("COMMAND_TIMEOUT_MS", "50")) * time.Millisecond,
		},
		RedisConfig: RedisConfig{
			ConnectionString: fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", ""), getEnv("REDIS_PORT", "")),
			Password:         getEnv("REDIS_PASSWORD", ""),
			Database:         getEnvAsInt("REDIS_DB", "0"),
			Size:             getEnvAsInt("REDIS_CACHE_SIZE_MB", "10"),
			Algorithm:        strings.ToLower(getEnv("REDIS_CACHE_ALGORITHM", "LRU")),
		},
	}
}
