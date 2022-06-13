package probes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadiness(t *testing.T) {
	readinessChecker := NewReadinessChecker()
	err := readinessChecker.CheckConnection(context.Background())

	assert.Nil(t, err)
}
