package config

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func unsetEnv(t *testing.T, envVar string) func() {
	t.Log("unset Envs")
	env := os.Getenv(envVar)
	os.Unsetenv(envVar)

	return func() {
		t.Log("set Env")
		os.Setenv(envVar, env)
	}
}

func setupEnv(t *testing.T, envVar, value string) func() {
	t.Log("setup Envs")
	env := os.Getenv(envVar)
	os.Setenv(envVar, value)

	return func() {
		t.Log("unset Env")
		os.Setenv(envVar, env)
	}
}

func TestConfigHelpers(t *testing.T) {
	t.Run("getEnv", func(t *testing.T) {
		t.Run("should return the default value if env not set", func(t *testing.T) {
			defer unsetEnv(t, "SERVER_LOG_LEVEL")()
			got := getEnv("SERVER_LOG_LEVEL", "warn")

			assert.Equal(t, "warn", got)
		})
		t.Run("should panic if env and default value not set", func(t *testing.T) {
			defer unsetEnv(t, "SERVER_LOG_LEVEL")()

			assert.Panics(t, func() {
				getEnv("SERVER_LOG_LEVEL", "")
			})
		})
	})
	t.Run("getEnvAsLogLevel", func(t *testing.T) {
		t.Run("should default to warn if no logLevel set", func(t *testing.T) {
			defer unsetEnv(t, "SERVER_LOG_LEVEL")()
			got := getEnvAsLogLevel("SERVER_LOG_LEVEL")

			assert.Equal(t, zerolog.WarnLevel, got)
		})
		t.Run("should set logLevel from environment", func(t *testing.T) {
			defer setupEnv(t, "SERVER_LOG_LEVEL", "info")()
			got := getEnvAsLogLevel("SERVER_LOG_LEVEL")

			assert.Equal(t, zerolog.InfoLevel, got)
		})
	})
	t.Run("getEnvAsSlice", func(t *testing.T) {
		t.Run("should return a slice from env", func(t *testing.T) {
			defer setupEnv(t, "CACHE_HASH_QUERY_IGNORE", "test,case")()
			got := getEnvAsSlice("CACHE_HASH_QUERY_IGNORE")

			assert.Equal(t, []string{"test", "case"}, got)
		})
		t.Run("should panic if env value not set", func(t *testing.T) {
			defer unsetEnv(t, "CACHE_HASH_QUERY_IGNORE")()

			assert.Panics(t, func() {
				getEnvAsSlice("CACHE_HASH_QUERY_IGNORE")
			})
		})
	})
	t.Run("getEnvAsInt", func(t *testing.T) {
		t.Run("should return an int from env", func(t *testing.T) {
			defer setupEnv(t, "TEST", "10")()
			got := getEnvAsInt("TEST", "")

			assert.Equal(t, 10, got)
		})
		t.Run("should return default value if not found in env", func(t *testing.T) {
			defer unsetEnv(t, "TEST")()
			got := getEnvAsInt("TEST", "20")

			assert.Equal(t, 20, got)
		})
		t.Run("should panic if not in env and no default value", func(t *testing.T) {
			defer unsetEnv(t, "TEST")()

			assert.Panics(t, func() {
				getEnvAsInt("TEST", "")
			})
		})
	})
	t.Run("getEnvAsFloat", func(t *testing.T) {
		t.Run("should return a float from env", func(t *testing.T) {
			defer setupEnv(t, "TEST", "10")()
			got := getEnvAsFloat("TEST", "")

			assert.Equal(t, 10.0, got)
		})
		t.Run("should return default value if not found in env", func(t *testing.T) {
			defer unsetEnv(t, "TEST")()
			got := getEnvAsFloat("TEST", "20.0")

			assert.Equal(t, 20.0, got)
		})
		t.Run("should panic if not in env and no default value", func(t *testing.T) {
			defer unsetEnv(t, "TEST")()

			assert.Panics(t, func() {
				getEnvAsFloat("TEST", "")
			})
		})
	})
	t.Run("getEnvAsBool", func(t *testing.T) {
		t.Run("should return a bool from env", func(t *testing.T) {
			defer setupEnv(t, "TEST", "true")()
			got := getEnvAsBool("TEST", "")

			assert.Equal(t, true, got)
		})
		t.Run("should return default value if not found in env", func(t *testing.T) {
			defer unsetEnv(t, "TEST")()
			got := getEnvAsBool("TEST", "false")

			assert.Equal(t, false, got)
		})
		t.Run("should panic if not in env and no default value", func(t *testing.T) {
			defer unsetEnv(t, "TEST")()

			assert.Panics(t, func() {
				getEnvAsBool("TEST", "")
			})
		})
	})
	t.Run("getEnvAsURL", func(t *testing.T) {
		t.Run("should return an URL from env", func(t *testing.T) {
			defer setupEnv(t, "TEST", "http://test.is")()
			got := getEnvAsURL("TEST", "")

			assert.Equal(t, "test.is", got.Host)
		})
		t.Run("should return default value as URL if not found in env", func(t *testing.T) {
			defer unsetEnv(t, "TEST")()
			got := getEnvAsURL("TEST", "http://test.de")

			assert.Equal(t, "test.de", got.Host)
		})
		t.Run("should panic if not found in env and no default value", func(t *testing.T) {
			defer unsetEnv(t, "TEST")()

			assert.Panics(t, func() {
				getEnvAsURL("TEST", "")
			})
		})
	})
}
