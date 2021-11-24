package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/skerkour/rz"
)

var (
	logLevel = map[string]rz.LogLevel{
		"debug":    rz.DebugLevel,
		"info":     rz.InfoLevel,
		"warn":     rz.WarnLevel,
		"error":    rz.ErrorLevel,
		"fatal":    rz.FatalLevel,
		"panic":    rz.PanicLevel,
		"none":     rz.NoLevel,
		"disabled": rz.Disabled,
	}
)

type RedisConfig struct {
	ConnectionString          string
	Password                  string
	Database                  int
	Size                      int
	Algorithm                 string
	CommandTimeoutMillisecond time.Duration
}

type MemoryConfig struct {
	Size      int
	Algorithm string
}

type ServerConfig struct {
	Port           string
	DownstreamHost string
	Storage        string
	LogLevel       rz.LogLevel
}

type CacheConfig struct {
	Strategy             string
	StaleWhileRevalidate int
	IgnorePaths          []string
	HashShouldQuery      bool
	HashQueryIgnore      map[string]bool
}
type Config struct {
	ServerConfig ServerConfig
	CacheConfig  CacheConfig
	RedisConfig  *RedisConfig
	MemoryConfig *MemoryConfig
}

func New() Config {
	serverConfig := ServerConfig{
		Port:           getEnv("SERVER_PORT", "8000"),
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
				HashShouldQuery:      getEnvAsBool("HASH_QUERY", ""),
				HashQueryIgnore:      hashQueryIgnoreMap(getEnvAsSlice("CACHE_IGNORE_ENDPOINTS")),
			},
			MemoryConfig: &MemoryConfig{
				Size:      getEnvAsInt("REDIS_CACHE_SIZE_MB", "50"),
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
			HashShouldQuery:      getEnvAsBool("HASH_QUERY", ""),
			HashQueryIgnore:      hashQueryIgnoreMap(getEnvAsSlice("CACHE_IGNORE_ENDPOINTS")),
		},
		RedisConfig: &RedisConfig{
			ConnectionString:          fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", ""), getEnv("REDIS_PORT", "")),
			Password:                  getEnv("REDIS_PASSWORD", ""),
			Database:                  getEnvAsInt("REDIS_DB", "0"),
			Size:                      getEnvAsInt("REDIS_CACHE_SIZE_MB", "10"),
			CommandTimeoutMillisecond: time.Duration(getEnvAsInt("REDIS_COMMAND_TIMEOUT_MS", "50")),
			Algorithm:                 strings.ToLower(getEnv("REDIS_CACHE_ALGORITHM", "LRU")),
		},
	}

}
