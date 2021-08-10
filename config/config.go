package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type RedisConfig struct {
	ConnectionString string
	Password         string
	Database         int
}

type ServerConfig struct {
	FailureModeDeny  bool
	Port             string
	DownstreamHost   string
	AllowedEndpoints []string
	IndempotencyKeys []string
}
type Config struct {
	RedisConfig  RedisConfig
	ServerConfig ServerConfig
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	if defaultVal == "" {
		log.Panic(key + " not set environment variable")
	}

	return defaultVal
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
			log.Fatalf("Key: %v not an int. DefaultValue: %v also not an int", key, defaultVal)
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
			log.Fatalf("Key: %v not a bool. DefaultValue: %v also not a bool", key, defaultVal)
		}
	}

	return value
}

func New() *Config {
	return &Config{
		ServerConfig: ServerConfig{
			Port:             getEnv("SERVER_PORT", "8000"),
			DownstreamHost:   getEnv("DOWNSTREAM_HOST", ""),
			AllowedEndpoints: getEnvAsSlice("DOWNSTREAM_ALLOWED_ENDPOINTS"),
			IndempotencyKeys: getEnvAsSlice("INDEMPOTENCY_KEYS"),
			FailureModeDeny:  getEnvAsBool("FAILURE_MODE_DENY", ""),
		},
		RedisConfig: RedisConfig{
			ConnectionString: fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", ""), getEnv("REDIS_PORT", "")),
			Password:         getEnv("REDIS_PASSWORD", ""),
			Database:         getEnvAsInt("REDIS_DB", "0"),
		},
	}
}
