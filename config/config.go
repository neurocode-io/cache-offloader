package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/skerkour/rz"
	"github.com/skerkour/rz/log"
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
	ExpirationTimeMinutes     time.Duration
	CommandTimeoutMillisecond time.Duration
}

type ServerConfig struct {
	FailureModeDeny bool
	Port            string
	DownstreamHost  string
	LogLevel        rz.LogLevel
}

type CacheConfig struct {
	Strategy             string
	Size                 int
	StaleWhileRevalidate int
	IgnorePaths          []string
	HashShouldQuery      bool
	HashQueryIgnore      map[string]bool
}
type Config struct {
	RedisConfig  RedisConfig
	ServerConfig ServerConfig
	CacheConfig  CacheConfig
}

func hashQueryIgnoreMap(queryIgnore []string) map[string]bool {
	hashQueryIgnoreMap := make(map[string]bool)

	for i := 0; i < len(queryIgnore); i++ {
		hashQueryIgnoreMap[queryIgnore[i]] = true
	}

	return hashQueryIgnoreMap
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	if defaultVal == "" {
		log.Fatal(key + " not set environment variable")
	}

	return defaultVal
}

func getEnvAsLogLevel(key string) rz.LogLevel {
	value, exists := os.LookupEnv(key)

	if !exists {
		log.Info("SERVER_LOG_LEVEL was not set, falling back to warn level")
		return rz.WarnLevel
	}

	if level, ok := logLevel[strings.ToLower(value)]; ok {
		return level
	}

	log.Warn(fmt.Sprintf("SERVER_LOG_LEVEL: %s is unknown, falling back to warn level", value))
	return rz.WarnLevel
}

func getEnvAsSlice(key string) []string {
	strSlice, _ := os.LookupEnv(key)
	if strSlice == "" {
		return nil
	}

	return strings.Split(strSlice, ",")
}

func getEnvAsInt(key, defaultVal string) int {
	valueStr := getEnv(key, defaultVal)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		value, err = strconv.Atoi(defaultVal)
		if err != nil {
			log.Fatal(fmt.Sprintf("Key: %v not an int. DefaultValue: %v also not an int", key, defaultVal))
		}
	}

	return value
}

func getEnvAsBool(key, defaultVal string) bool {
	valueStr := getEnv(key, defaultVal)
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		value, err = strconv.ParseBool(defaultVal)
		if err != nil {
			log.Fatal(fmt.Sprintf("Key: %v not a bool. DefaultValue: %v also not a bool", key, defaultVal))
		}
	}

	return value
}

func New() *Config {
	return &Config{
		ServerConfig: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8000"),
			DownstreamHost:  getEnv("DOWNSTREAM_HOST", ""),
			LogLevel:        getEnvAsLogLevel("SERVER_LOG_LEVEL"),
			FailureModeDeny: getEnvAsBool("FAILURE_MODE_DENY", ""),
		},
		CacheConfig: CacheConfig{
			Strategy:             getEnv("CACHE_STRATEGY", ""),
			IgnorePaths:          getEnvAsSlice("CACHE_IGNORE_ENDPOINTS"),
			Size:                 getEnvAsInt("CACHE_SIZE_MB", "10"),
			StaleWhileRevalidate: getEnvAsInt("CACHE_STALE_WHILE_REVALIDATE_SEC", "5"),
			HashShouldQuery:      getEnvAsBool("HASH_QUERY", ""),
			HashQueryIgnore:      hashQueryIgnoreMap(getEnvAsSlice("CACHE_IGNORE_ENDPOINTS")),
		},
		RedisConfig: RedisConfig{
			ConnectionString:          fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", ""), getEnv("REDIS_PORT", "")),
			Password:                  getEnv("REDIS_PASSWORD", ""),
			Database:                  getEnvAsInt("REDIS_DB", "0"),
			CommandTimeoutMillisecond: time.Duration(getEnvAsInt("REDIS_COMMAND_TIMEOUT", "50")),
			ExpirationTimeMinutes:     time.Duration(getEnvAsInt("REDIS_EXPIRATION_TIME_MINUTES", "5")),
		},
	}
}
