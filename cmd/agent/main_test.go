package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	name := "test saveMetricss"

	t.Run(name, func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())
		err := saveMetrics(ctx)
		require.NoError(t, err)
		time.AfterFunc(4*time.Second, cancel)
	})
}

func TestGet(t *testing.T) {
	name := "test getting metrics"
	mm := make(map[string]interface{})

	t.Run(name, func(t *testing.T) {
		m := M.Get()
		assert.Equal(t, m, mm)

	})

}
