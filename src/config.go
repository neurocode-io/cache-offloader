package main

import (
	"log"
	"os"
	"strings"
)

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
