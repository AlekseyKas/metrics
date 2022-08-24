package helpers

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/server/handlers"
	"go.uber.org/zap"
)

// Load metrics from file storage.
func LoadFromFile(wg *sync.WaitGroup, logger *zap.Logger, env config.Args) error {
	if env.Restore && fileExist(env.StoreFile) {
		file, err := os.ReadFile(env.StoreFile)
		if err != nil {
			logger.Error("Error open file for writing: ", zap.Error(err))
			wg.Done()
			return err
		}
		handlers.StorageM.LoadMetricsFile(file)
	}
	return nil
}

// Checking exist file or don't exist.
func fileExist(file string) bool {
	var b bool
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return b
	}
	b = true
	return b
}

// Wait siglans SIGTERM, SIGINT, SIGQUIT.
func WaitSignals(cancel context.CancelFunc, logger *zap.Logger, wg *sync.WaitGroup) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		sig := <-terminate
		switch sig {
		case os.Interrupt:
			if config.ArgsM.DBURL != "" {
				handlers.StorageM.StopDB()
			}
			cancel()
			wg.Done()
			logger.Info("Terminate signal OS!")
			return
		}
	}
}

// Sync metrics with file storage.
func SyncFile(ctx context.Context, wg *sync.WaitGroup, logger *zap.Logger, env config.Args) {
	if env.StoreFile == "" {
		for {
			<-ctx.Done()
			logger.Info("File syncing is down")
			wg.Done()
			return
		}
	} else {
		if env.StoreInterval == 0 {
			metrics, err := handlers.StorageM.GetMetricsJSON()
			if err != nil {
				logger.Error("Error getting metrics format JSON GetMetricsJSON: ", zap.Error(err))
			}
			file, err := os.Create(env.StoreFile)
			if err != nil {
				logger.Error("Error open file for writing: ", zap.Error(err))
			}
			defer file.Close()

			data, err := json.Marshal(metrics)
			if err != nil {
				logger.Error("Error marshaling metrics : ", zap.Error(err))
			}
			_, err = file.Write(data)
			if err != nil {
				logger.Error("Error write data to file: ", zap.Error(err))
			}
			for {
				<-ctx.Done()
				logger.Info("File syncing is down")
				wg.Done()
				return
			}
		} else {
			for {
				select {
				case <-ctx.Done():
					logger.Info("File syncing is down")
					wg.Done()
					return
				case <-time.After(env.StoreInterval):
					metrics, _ := handlers.StorageM.GetMetricsJSON()
					file, err := os.OpenFile(env.StoreFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
					if err != nil {
						logger.Error("Error open file for writing: ", zap.Error(err))
					}
					defer file.Close()

					data, err := json.Marshal(metrics)
					if err != nil {
						logger.Error("Error marshaling metrics : ", zap.Error(err))
					}
					_, err = file.Write(data)
					if err != nil {
						logger.Error("Error writing data to file: ", zap.Error(err))
					}
				}
			}
		}
	}
}
