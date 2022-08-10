package main

import (
	"context"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/stretchr/testify/require"
)

// s := &storage.MetricsStore{
// 	MM: structs.Map(storage.Metrics{}),
// }

// termEnvFlags()
// handlers.SetStorage(s)

// func Test_termEnvFlags(t *testing.T) {

// 	termEnvFlags()

// 	t.Run("Test terminate flags", func(t *testing.T) {
// 		require.Equal(t, config.ArgsM.Address, "127.0.0.1:8080")
// 		require.Equal(t, config.ArgsM.StoreFile, "")
// 		require.Equal(t, config.ArgsM.DBURL, "")
// 		require.Equal(t, config.ArgsM.StoreInterval, time.Duration(300000000000))
// 	})
// }

// func Test_loadFromFile(t *testing.T) {

// 	t.Run("Test loadFromFile", func(t *testing.T) {
// 		err := loadFromFile(config.ArgsM)
// 		require.NoError(t, err)
// 	})
// }

func Test_syncFile(t *testing.T) {
	var wg sync.WaitGroup
	t.Run("Test syncFile", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
		wg.Add(1)
		go syncFile(config.ArgsM, ctx)
		wg.Add(1)
		go waitSignals(cancel)
	})
}

// func Test_waitSignals(t *testing.T) {
// 	var wg sync.WaitGroup
// 	logrus.Info("yyyyyyyyyyyyyyyyyyyyyy")
// 	_, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*3))
// 	logrus.Info("ppppppppppppppppppppppp")
// 	t.Run("Test waitSignals 3 seconds", func(t *testing.T) {
// 		wg.Add(1)
// 		waitSignals(cancel)
// 	})

// }

func Test_fileExist(t *testing.T) {

	tests := []struct {
		name string
		file string
		want bool
	}{
		{
			name: "file don't exist",
			file: ".test",
			want: false,
		},
		{
			name: "file exist",
			file: ".main.go",
			want: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoFileExists(t, tt.file)
		})
	}
}
