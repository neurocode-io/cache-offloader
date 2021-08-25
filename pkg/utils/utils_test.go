package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type endpointTests []struct {
	in       string
	expected bool
}

func TestAllowedEndoints(t *testing.T) {
	allowedEndpoints := []string{"/management/loggers/*", "/api-docs/*"}

	assert.Equal(t, VariableMatchesRegexIn("/management/loggers/configuration", allowedEndpoints), true)
	assert.Equal(t, VariableMatchesRegexIn("/management/loggersbb/configuration", allowedEndpoints), false)
	assert.Equal(t, VariableMatchesRegexIn("/api-docs/t/a", allowedEndpoints), true)
	assert.Equal(t, VariableMatchesRegexIn("/management/health", allowedEndpoints), false)

}
