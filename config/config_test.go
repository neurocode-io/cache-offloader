package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("inMemory config", func(t *testing.T) {
		defer setupEnv(t, "SERVER_STORAGE", "memory")()
		config := New()

		assert.NotNil(t, config)
		assert.Equal(t, "", config.RedisConfig.ConnectionString)
	})
	t.Run("redis config", func(t *testing.T) {
		defer setupEnv(t, "SERVER_STORAGE", "redis")()
		defer setupEnv(t, "REDIS_CONNECTION_STRING", "redis")()
		config := New()

		assert.NotNil(t, config)
		assert.Equal(t, "redis", config.RedisConfig.ConnectionString)
	})
}
