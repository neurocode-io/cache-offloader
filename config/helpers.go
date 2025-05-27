package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func hashQueryIgnoreMap(queryIgnore []string) map[string]bool {
	hashQueryIgnoreMap := make(map[string]bool)
	for _, q := range queryIgnore {
		hashQueryIgnoreMap[strings.ToLower(strings.TrimSpace(q))] = true
	}
	return hashQueryIgnoreMap
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}

	if defaultVal == "" {
		log.Panic().Msgf("%s is not set", key)
	}

	return defaultVal
}

func getEnvAsLogLevel(key string) zerolog.Level {
	value, exists := os.LookupEnv(key)

	if !exists {
		log.Info().Msg("LOG_LEVEL was not set, falling back to warn level")

		return zerolog.WarnLevel
	}

	if level, ok := logLevel[strings.ToLower(value)]; ok {
		return level
	}

	log.Warn().Msgf("LOG_LEVEL: %s is unknown, falling back to warn level", value)

	return zerolog.WarnLevel
}

func getEnvAsSlice(key string) []string {
	strSlice, _ := os.LookupEnv(key)
	if strSlice == "" {
		return []string{}
	}
	parts := strings.Split(strSlice, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

func getEnvAsInt(key, defaultVal string) int {
	valueStr := getEnv(key, defaultVal)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		value, err = strconv.Atoi(defaultVal)
		if err != nil {
			log.Panic().Msgf("Key: %v not an int. DefaultValue: %v also not an int", key, defaultVal)
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
			log.Panic().Msgf("Key: %v not an int. DefaultValue: %v also not an int", key, defaultVal)
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
			log.Panic().Msgf("Key: %v not a bool. DefaultValue: %v also not a bool", key, defaultVal)
		}
	}

	return value
}

func getEnvAsURL(key, defaultVal string) *url.URL {
	valueStr := getEnv(key, defaultVal)
	downstreamURL, err := url.Parse(valueStr)
	if err != nil {
		log.Panic().Msgf("Could not parse downstream url: %s", downstreamURL)
	}

	return downstreamURL
}
