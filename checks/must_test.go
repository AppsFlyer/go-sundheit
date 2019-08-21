package checks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Must(&CustomCheck{}, nil)
		}, "Must should not panic when check creation succeeds")
	})

	t.Run("panic", func(t *testing.T) {
		err := errors.New("failed")
		assert.PanicsWithValue(t, err, func() {
			Must(nil, err)
		}, "Must should panic when check creation fails")
	})
}
