package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/fatih/structs"
	"go.uber.org/zap"

	"github.com/AlekseyKas/metrics/internal/agent/helpers"
	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
)

var (
	// Build version.
	buildVersion string = "N/A"

	// Build date.
	buildDate string = "N/A"

	// Build commit.
	buildCommit string = "N/A"
)

func main() {
	// Init waitGroup.
	var wg = &sync.WaitGroup{}
	// Init storage of metrics.
	var storageM storage.StorageAgent
	// Init loger zap
	logger, err := zap.NewProduction()
	if err != nil {
		logger.Error("Error init logger: ", zap.Error(err))
	}

	// Init map of metrics.
	var MapMetrics = structs.Map(storage.Metrics{})
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	storageM = s
	// Tegminate enviranment and flags.
	config.TermEnvFlagsAgent()
	// Init logger.
	storage.InitLogger(logger)
	// Init context with cancel.
	ctx, cancel := context.WithCancel(context.Background())
	// Add count waitgroup.
	wg.Add(4)
	// Wait signal from operation system.
	go helpers.WaitSignals(cancel, logger, wg)
	// Update metrics terminating.
	go helpers.UpdateMetrics(ctx, config.ArgsM.PollInterval, wg, logger, storageM)
	// Update new metrics.
	go helpers.UpdateMetricsNew(ctx, config.ArgsM.PollInterval, wg, logger, storageM)
	// Send metrics to server.
	go helpers.SendMetrics(ctx, wg, logger, ArgsM.PubKey, storageM)

	// Printing build options.
	fmt.Printf("Build version:%s \n", buildVersion)
	fmt.Printf("Build date:%s \n", buildDate)
	fmt.Printf("Build commit:%s \n", buildCommit)
	// Init waiting.
	wg.Wait()
}
