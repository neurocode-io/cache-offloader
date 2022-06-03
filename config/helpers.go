package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/skerkour/rz"
	"github.com/skerkour/rz/log"
)

func hashQueryIgnoreMap(queryIgnore []string) map[string]bool {
	hashQueryIgnoreMap := make(map[string]bool)

	for i := 0; i < len(queryIgnore); i++ {
		hashQueryIgnoreMap[queryIgnore[i]] = true
	}

	return hashQueryIgnoreMap
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
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

func getEnvAsFloat(key, defaultVal string) float64 {
	valueStr := getEnv(key, defaultVal)
	bitSize := 64
	value, err := strconv.ParseFloat(valueStr, bitSize)
	if err != nil {
		value, err = strconv.ParseFloat(defaultVal, bitSize)
		if err != nil {
			log.Fatal(fmt.Sprintf("Key: %v not an int. DefaultValue: %v also not an int", key, defaultVal))
		}
	}

	return value
}

func getEnvAsBool(key, defaultVal string) bool {
	valueStr := strings.Trim(getEnv(key, defaultVal), "\"")
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		value, err = strconv.ParseBool(defaultVal)
		if err != nil {
			log.Fatal(fmt.Sprintf("Key: %v not a bool. DefaultValue: %v also not a bool", key, defaultVal))
		}
	}

	return value
}
