package helpers

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	key := "key_test"
	storageM = s
	config.TermEnvFlagsAgent()
	logger, _ := zap.NewProduction()
	t.Run("SendMetrics", func(t *testing.T) {
		wg.Add(2)
		go SendMetrics(ctx, wg, logger, key, storageM)
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
	// key := []byte("-----BEGIN PRIVATE KEY-----\n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\nbJzm2yr4DcY5MUeRTtG06l5t\n-----END PRIVATE KEY-----\n")
	key := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDlITF08iApaazMvatDaC/oKxQ2xsdv7UjsTBIIshx/soyeXtK0QjXJUiBl1iqCrwSbg3gHOnTzkH8bO1r5WywNCj40dvmiNOotUfb28s4P8mawkpP4JvE0lIU5h0V97y9sSsQazyqd1jqMelMPp4CYSOrhbTcIVieku/RQXV42tU+AUD470fiAlhLEXCAVm+8U5cI8bStSCyEJLxLJSicxBXQnPJ6Z7UwIxdEI0XHRUyu9/LMXa3qZFKo3ad5Fgt6uveI9BmNK1yYOk3h09fBW37ip8eOFHS1aY82V/Qwak4KJ5BO4YxYVxlaAJ8mpu4EkdSp0unr8qMsQM0h4P51R+9EDyYuP2QXQD9ch/x1P83HNQAj+XGKPSGOF7IZkbA3p+fv3ZwcSMfg3Fmqym96utW3DKPovAt+r/430uQvwfjq0UTvsOU0s80ObhYbfUfq2708obLj02dD7HjmHWnSKfwpWcyS8XW+TuCU6uEpp5yODGpy5ArveF4tiKoBHIqU=")
	storageM = s
	ctx, cancel := context.WithCancel(context.Background())
	logger, _ := zap.NewProduction()
	t.Run("sendMetricsSlice", func(t *testing.T) {
		err := os.WriteFile("key", key, 0600)
		require.NoError(t, err)
		wg.Add(1)
		err = SendMetricsSlice(ctx, logger, config.ArgsM.Address, "key", []byte(config.ArgsM.Key), storageM)
		require.Error(t, err)
		time.Sleep(time.Second * 2)
		cancel()
		wg.Done()
	})
}

func TestUpdateMetrics(t *testing.T) {
	tests := []struct {
		name         string
		pollInterval time.Duration
	}{
		{
			name:         "Pollinterval = 1s",
			pollInterval: 1,
		},
	}
	var storageM storage.StorageAgent
	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	storageM = s
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger, _ := zap.NewProduction()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg.Add(3)
			go UpdateMetrics(ctx, tt.pollInterval, wg, logger, storageM)
			go UpdateMetricsNew(ctx, tt.pollInterval, wg, logger, storageM)
			go WaitSignals(cancel, logger, wg)
			time.Sleep(time.Second * 4)
			wg.Done()
			defer cancel()
		})
	}
}
