package helpers

import (
	"context"
	_ "net/http/pprof"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/server/handlers"
	"github.com/AlekseyKas/metrics/internal/storage"
)

func Test_syncFile(t *testing.T) {
	f, _ := os.CreateTemp("/tmp/", "file")

	var wg = &sync.WaitGroup{}
	tests := []struct {
		name   string
		config config.Args
	}{
		{
			name: "first",
			config: config.Args{
				StoreFile:     "",
				StoreInterval: 1,
			},
		},
		{
			name: "first",
			config: config.Args{
				StoreFile:     "",
				StoreInterval: 0,
			},
		},

		{
			name: "Second",
			config: config.Args{
				StoreFile:     f.Name(),
				StoreInterval: 1,
			},
		},
		{
			name: "Third",
			config: config.Args{
				StoreFile:     f.Name(),
				StoreInterval: 0,
			},
		},
	}
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	handlers.SetStorage(s)
	logger, _ := zap.NewProduction()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			wg.Add(1)
			go SyncFile(ctx, wg, logger, tt.config)
			if tt.config.StoreFile != "" {
				require.FileExists(t, tt.config.StoreFile)
			} else {
				require.NoFileExists(t, tt.config.StoreFile)
			}
			wg.Add(1)
			go WaitSignals(cancel, logger, wg)
			time.Sleep(time.Second * 2)
		})
	}
}

func Test_LoadFromFile(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Args
		wantErr bool
	}{
		{
			name: "first",
			config: config.Args{
				StoreFile:     "file",
				StoreInterval: 1,
				Restore:       true,
			},
		},
		{
			name: "second",
			config: config.Args{
				StoreFile:     "",
				StoreInterval: 1,
				Restore:       true,
			},
		},
		{
			name: "third",
			config: config.Args{
				StoreFile:     "",
				StoreInterval: 1,
				Restore:       false,
			},
		},
		{
			name: "4th",
			config: config.Args{
				StoreFile:     "file",
				StoreInterval: 1,
				Restore:       false,
			},
		},
	}
	logger, _ := zap.NewProduction()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("/tmp/", tt.config.StoreFile)
			require.FileExists(t, f.Name())
			require.NoError(t, err)
			err = LoadFromFile(logger, tt.config)
			require.NoError(t, err)
		})
	}
}

func Test_fileExist(t *testing.T) {
	f, _ := os.CreateTemp("/tmp/", "file")
	tests := []struct {
		name    string
		config  config.Args
		wantErr bool
	}{
		{
			name: "first",
			config: config.Args{
				StoreFile:     f.Name(),
				StoreInterval: 1,
				Restore:       true,
			},
		},
		{
			name: "second",
			config: config.Args{
				StoreFile:     "",
				StoreInterval: 1,
				Restore:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := fileExist(tt.config.StoreFile)
			if b {
				require.True(t, b)
			} else {
				require.False(t, b)
			}
		})
	}
}