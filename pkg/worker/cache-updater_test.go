package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheUpdater(t *testing.T) {
	t.Run("UpdateQueue shouldnt panic on negative numbers", func(t *testing.T) {
		q := NewUpdateQueue(-1)

		assert.NotNil(t, q)
	})
	t.Run("should do the work in a function", func(t *testing.T) {
		q := NewUpdateQueue(1)
		q.Start("test2", func() {
			t.Log("test work")
		})
	})
}
