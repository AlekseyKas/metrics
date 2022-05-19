package main

import (
	"context"
	"testing"
	"time"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
)

var Hash string

func TestClient(t *testing.T) {
	name := "test saveMetricss"
	var storageM storage.StorageAgent
	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	storageM = s
	t.Run(name, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		p := config.LoadConfig()
		err := sendMetricsSlice(ctx, p.Address, []byte(p.Key), storageM)
		require.Error(t, err)
		time.AfterFunc(4*time.Second, cancel)
	})
}

func TestSaveHash(t *testing.T) {
	f := float64(45)
	jm := storage.JSONMetrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &f,
	}

	key := "key"
	type args struct {
		JSONMetric *storage.JSONMetrics
		key        []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "first",
			args: args{
				JSONMetric: &jm,
				key:        []byte(key),
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if hash, err := SaveHash(tt.args.JSONMetric, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("SaveHash() error = %v, wantErr %v", err, tt.wantErr)
				assert.Empty(t, hash)
			}

		})
	}
}
