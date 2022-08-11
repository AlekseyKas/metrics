package main

import (
	"context"
	"io/ioutil"
	_ "net/http/pprof"
	"testing"
	"time"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
)

func Test_syncFile(t *testing.T) {
	f, _ := ioutil.TempFile("/tmp/", "file")

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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			wg.Add(1)
			go syncFile(tt.config, ctx)
			if tt.config.StoreFile != "" {
				require.FileExists(t, tt.config.StoreFile)
			} else {
				require.NoFileExists(t, tt.config.StoreFile)
			}
			wg.Add(1)
			time.Sleep(time.Second * 2)
			cancel()
			wg.Done()
		})
	}
}

func Test_loadFromFile(t *testing.T) {

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
	for _, tt := range tests {
		_, err := ioutil.TempFile("/tmp/", tt.config.StoreFile)
		require.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			err = loadFromFile(tt.config)
			require.NoError(t, err)
		})
	}
}

func Test_fileExist(t *testing.T) {
	f, _ := ioutil.TempFile("/tmp/", "file")
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
