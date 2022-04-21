package main

import (
	"context"
	"testing"
	"time"

	"github.com/AlekseyKas/metrics/internal/config"
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
	// require.NoError(t, err)
	t.Run(name, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		p := config.LoadConfig()
		err := sendMetricsJSON(ctx, p.Address)
		require.Error(t, err)
		time.AfterFunc(4*time.Second, cancel)
	})
}
