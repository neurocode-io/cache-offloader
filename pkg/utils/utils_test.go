package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowedEndoints(t *testing.T) {
	allowedEndpoints := []string{"/management/loggers/", "/api-docs"}

	assert.Equal(t, true, VariableMatchesRegexIn("/management/loggers/configuration", allowedEndpoints))
	assert.Equal(t, true, VariableMatchesRegexIn("/management/loggersbb/configuration", allowedEndpoints))
	assert.Equal(t, true, VariableMatchesRegexIn("/api-docs/t/a", allowedEndpoints))
	assert.Equal(t, true, VariableMatchesRegexIn("/api-docs/t/a/", allowedEndpoints))
	assert.Equal(t, false, VariableMatchesRegexIn("/management/health", allowedEndpoints))
	assert.Equal(t, true, VariableMatchesRegexIn("/management/loggers", allowedEndpoints))
	assert.Equal(t, true, VariableMatchesRegexIn("/management/loggers/", allowedEndpoints))
}
