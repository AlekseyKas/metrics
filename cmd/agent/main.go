package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/fatih/structs"

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
	// Init map of metrics.
	var MapMetrics = structs.Map(storage.Metrics{})
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	storageM = s
	// Tegminate enviranment and flags.
	config.TermEnvFlagsAgent()
	// Init context with cancel.
	ctx, cancel := context.WithCancel(context.Background())
	// Add count waitgroup.
	wg.Add(4)
	// Wait signal from operation system.
	go helpers.WaitSignals(cancel, wg)
	// Update metrics terminating.
	go helpers.UpdateMetrics(ctx, config.ArgsM.PollInterval, wg, storageM)
	// Update new metrics.
	go helpers.UpdateMetricsNew(ctx, config.ArgsM.PollInterval, wg, storageM)
	// Send metrics to server.
	go helpers.SendMetrics(ctx, wg, storageM)

	// Printing build options.
	fmt.Printf("Build version:%s \n", buildVersion)
	fmt.Printf("Build date:%s \n", buildDate)
	fmt.Printf("Build commit:%s \n", buildCommit)
	// Init waiting.
	wg.Wait()
}
