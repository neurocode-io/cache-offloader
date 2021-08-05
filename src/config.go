package main

import (
	"log"
	"os"
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
