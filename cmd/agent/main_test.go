package main

import (
	"context"
	"testing"
	"time"

	"github.com/AlekseyKas/metrics/cmd/agent/helpers"
	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/require"
)

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
		err := helpers.SendMetricsSlice(ctx, p.Address, []byte(p.Key), storageM)
		require.Error(t, err)
		time.AfterFunc(4*time.Second, cancel)
	})
}

// func TestSaveHash(t *testing.T) {
// 	f := float64(45)
// 	c := int64(4)

// 	type args struct {
// 		JSONMetric *storage.JSONMetrics
// 		key        []byte
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		sha  string
// 	}{
// 		{
// 			name: "first",
// 			args: args{
// 				JSONMetric: &storage.JSONMetrics{
// 					ID:    "Alloc",
// 					MType: "gauge",
// 					Value: &f,
// 				},
// 				key: []byte("key"),
// 			},
// 			sha: "f5e9ca6c3337abf049e8199a895fcbe3468c7f2c33d0126546e698976418f27e",
// 		},
// 		{
// 			name: "first",
// 			args: args{
// 				JSONMetric: &storage.JSONMetrics{
// 					ID:    "Pollcount",
// 					MType: "counter",
// 					Delta: &c,
// 				},
// 				key: []byte("key111"),
// 			},
// 			sha: "af087c9d1c0119ccb77efa66efc24250f9e515d665c925690d7f1c27d3f5c88a",
// 		},
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := saveHash(tt.args.JSONMetric, tt.args.key)
// 			require.Empty(t, err)
// 			require.Equal(t, tt.sha, tt.args.JSONMetric.Hash)
// 		})
// 	}
// }

// func TestSendMetrics(t *testing.T) {
// 	var wg = &sync.WaitGroup{}
// 	var storageM storage.StorageAgent
// 	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
// 	ctx, cancel := context.WithCancel(context.Background())
// 	s := &storage.MetricsStore{
// 		MM: MapMetrics,
// 	}
// 	storageM = s
// 	termEnvFlags()
// 	t.Run("SendMetrics", func(t *testing.T) {
// 		wg.Add(2)
// 		go SendMetrics(ctx, wg, storageM)
// 		time.Sleep(time.Second * 2)
// 		cancel()
// 		wg.Done()
// 	})

// }

// func Test_sendMetricsSlice(t *testing.T) {
// 	var wg = &sync.WaitGroup{}
// 	var storageM storage.StorageAgent
// 	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
// 	s := &storage.MetricsStore{
// 		MM: MapMetrics,
// 	}
// 	storageM = s
// 	ctx, cancel := context.WithCancel(context.Background())
// 	t.Run("sendMetricsSlice", func(t *testing.T) {
// 		wg.Add(1)
// 		sendMetricsSlice(ctx, config.ArgsM.Address, []byte(config.ArgsM.Key), storageM)
// 		time.Sleep(time.Second * 2)
// 		cancel()
// 		wg.Done()
// 	})
// }
