package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseModel(t *testing.T) {
	t.Run("Stale", func(t *testing.T) {
		t.Run("returns true when stale", func(t *testing.T) {
			r := Response{StaleValue: StaleValue}
			assert.True(t, r.IsStale())
		})
		t.Run("returns false when not stale", func(t *testing.T) {
			r := Response{StaleValue: FreshValue}
			assert.False(t, r.IsStale())
		})
	})

	t.Run("Response serialization", func(t *testing.T) {
		t.Run("StaleValue property should not be JSON serialized", func(t *testing.T) {
			r := Response{
				Header: map[string][]string{
					"foo": {"bar"},
				},
				Body:       []byte("body"),
				Status:     200,
				StaleValue: StaleValue,
			}

			serialized, err := json.Marshal(r)
			assert.Nil(t, err)
			assert.NotContains(t, string(serialized), "StaleValue")
		})
	})
}
