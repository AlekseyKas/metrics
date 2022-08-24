package helpers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fatih/structs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
)

func TestSaveHash(t *testing.T) {
	f := float64(45)
	c := int64(4)

	type args struct {
		JSONMetric *storage.JSONMetrics
		key        []byte
	}
	tests := []struct {
		name string
		sha  string
		args args
	}{
		{
			name: "first",
			args: args{
				JSONMetric: &storage.JSONMetrics{
					ID:    "Alloc",
					MType: "gauge",
					Value: &f,
				},
				key: []byte("key"),
			},
			sha: "f5e9ca6c3337abf049e8199a895fcbe3468c7f2c33d0126546e698976418f27e",
		},
		{
			name: "first",
			args: args{
				JSONMetric: &storage.JSONMetrics{
					ID:    "Pollcount",
					MType: "counter",
					Delta: &c,
				},
				key: []byte("key111"),
			},
			sha: "af087c9d1c0119ccb77efa66efc24250f9e515d665c925690d7f1c27d3f5c88a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := saveHash(tt.args.JSONMetric, tt.args.key)
			require.Empty(t, err)
			require.Equal(t, tt.sha, tt.args.JSONMetric.Hash)
		})
	}
}

func TestSendMetrics(t *testing.T) {
	var wg = &sync.WaitGroup{}
	var storageM storage.StorageAgent
	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	ctx, cancel := context.WithCancel(context.Background())
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	storageM = s
	config.TermEnvFlagsAgent()
	logger, _ := zap.NewProduction()
	t.Run("SendMetrics", func(t *testing.T) {
		wg.Add(2)
		go SendMetrics(ctx, wg, logger, storageM)
		time.Sleep(time.Second * 2)
		cancel()
		wg.Done()
	})

}

func Test_sendMetricsSlice(t *testing.T) {
	var wg = &sync.WaitGroup{}
	var storageM storage.StorageAgent
	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	storageM = s
	ctx, cancel := context.WithCancel(context.Background())
	logger, _ := zap.NewProduction()
	t.Run("sendMetricsSlice", func(t *testing.T) {
		wg.Add(1)
		err := SendMetricsSlice(ctx, logger, config.ArgsM.Address, []byte(config.ArgsM.Key), storageM)
		if err != nil {
			logrus.Error("Error sending slice metrics: ", err)
		}
		time.Sleep(time.Second * 2)
		cancel()
		wg.Done()
	})
}
