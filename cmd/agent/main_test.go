package main

import (
	"context"
	"testing"
	"time"

	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	name := "test saveMetricss"
	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	SetStorageAgent(s)
	require.NoError(t, err)
	t.Run(name, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		err := sendMetricsJSON(ctx)
		require.NoError(t, err, "message")
		time.AfterFunc(4*time.Second, cancel)
	})
}

// func TestGet(t *testing.T) {
// 	name := "test getting metrics"
// 	mm := make(map[string]interface{})

// 	t.Run(name, func(t *testing.T) {
// 		m := M.Get()
// 		assert.Equal(t, m, mm)

// 	})

// }
