package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("inMemory config", func(t *testing.T) {
		defer setupEnv(t, "SERVER_STORAGE", "memory")()
		defer setupEnv(t, "CACHE_STRATEGY", "memory")()
		defer setupEnv(t, "DOWNSTREAM_HOST", "http://localhost:8080")()
		config := New()

		assert.NotNil(t, config)
		assert.Equal(t, "", config.RedisConfig.ConnectionString)
	})
	t.Run("redis config", func(t *testing.T) {
		defer setupEnv(t, "SERVER_STORAGE", "redis")()
		defer setupEnv(t, "REDIS_CONNECTION_STRING", "redis")()
		defer setupEnv(t, "CACHE_STRATEGY", "redis")()
		defer setupEnv(t, "DOWNSTREAM_HOST", "http://localhost:8080")()
		config := New()

		assert.NotNil(t, config)
		assert.Equal(t, "redis", config.RedisConfig.ConnectionString)
	})

	t.Run("header configuration", func(t *testing.T) {
		tests := []struct {
			name              string
			hashHeaders       string
			hashHeadersIgnore string
			wantHeaders       []string
		}{
			{
				name:              "empty headers",
				hashHeaders:       "",
				hashHeadersIgnore: "",
				wantHeaders:       []string{},
			},
			{
				name:              "single header",
				hashHeaders:       "Authorization",
				hashHeadersIgnore: "",
				wantHeaders:       []string{"Authorization"},
			},
			{
				name:              "multiple headers",
				hashHeaders:       "Authorization,X-User-ID,Accept",
				hashHeadersIgnore: "",
				wantHeaders:       []string{"Authorization", "X-User-ID", "Accept"},
			},
			{
				name:              "headers with ignore",
				hashHeaders:       "Authorization,X-User-ID,Accept",
				hashHeadersIgnore: "X-User-ID",
				wantHeaders:       []string{"Authorization", "X-User-ID", "Accept"},
			},
			{
				name:              "case insensitive headers",
				hashHeaders:       "AUTHORIZATION,x-user-id,accept",
				hashHeadersIgnore: "X-USER-ID",
				wantHeaders:       []string{"AUTHORIZATION", "x-user-id", "accept"},
			},
			{
				name:              "whitespace in headers",
				hashHeaders:       " Authorization , X-User-ID , Accept ",
				hashHeadersIgnore: " X-User-ID ",
				wantHeaders:       []string{"Authorization", "X-User-ID", "Accept"},
			},
			{
				name:              "duplicate headers",
				hashHeaders:       "Authorization,Authorization,X-User-ID",
				hashHeadersIgnore: "",
				wantHeaders:       []string{"Authorization", "Authorization", "X-User-ID"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				defer setupEnv(t, "CACHE_HASH_HEADERS", tt.hashHeaders)()
				defer setupEnv(t, "CACHE_HASH_HEADERS_IGNORE", tt.hashHeadersIgnore)()

				config := New()

				assert.Equal(t, tt.wantHeaders, config.CacheConfig.HashHeaders)
			})
		}
	})

	t.Run("query configuration", func(t *testing.T) {
		tests := []struct {
			name            string
			shouldHashQuery string
			queryIgnore     string
			wantHashQuery   bool
			wantIgnore      map[string]bool
		}{
			{
				name:            "default query hashing",
				shouldHashQuery: "",
				queryIgnore:     "",
				wantHashQuery:   true,
				wantIgnore:      map[string]bool{},
			},
			{
				name:            "disabled query hashing",
				shouldHashQuery: "false",
				queryIgnore:     "",
				wantHashQuery:   false,
				wantIgnore:      map[string]bool{},
			},
			{
				name:            "query ignore parameters",
				shouldHashQuery: "true",
				queryIgnore:     "timestamp,request_id",
				wantHashQuery:   true,
				wantIgnore:      map[string]bool{"timestamp": true, "request_id": true},
			},
			{
				name:            "whitespace in query ignore",
				shouldHashQuery: "true",
				queryIgnore:     " timestamp , request_id ",
				wantHashQuery:   true,
				wantIgnore:      map[string]bool{"timestamp": true, "request_id": true},
			},
			{
				name:            "case sensitive query ignore",
				shouldHashQuery: "true",
				queryIgnore:     "Timestamp,Request_ID",
				wantHashQuery:   true,
				wantIgnore:      map[string]bool{"timestamp": true, "request_id": true},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				defer setupEnv(t, "CACHE_SHOULD_HASH_QUERY", tt.shouldHashQuery)()
				defer setupEnv(t, "CACHE_HASH_QUERY_IGNORE", tt.queryIgnore)()

				config := New()

				assert.Equal(t, tt.wantHashQuery, config.CacheConfig.ShouldHashQuery)
				assert.Equal(t, tt.wantIgnore, config.CacheConfig.HashQueryIgnore)
			})
		}
	})
}
