package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/fatih/structs"
	"github.com/go-chi/chi"

	"go.uber.org/zap"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/server/handlers"
	"github.com/AlekseyKas/metrics/internal/server/helpers"
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
	// Init wait group.
	var wg sync.WaitGroup
	// Default context and cancel.
	ctx, cancel := context.WithCancel(context.Background())
	// Init loger zap
	logger, err := zap.NewProduction()
	if err != nil {
		logger.Error("Error init logger: ", zap.Error(err))
	}
	// Send logger to storage
	storage.InitLogger(logger)
	// Send logger to handlers
	handlers.InitLogger(logger)
	// Inint storage server.
	s := &storage.MetricsStore{
		MM:  structs.Map(storage.Metrics{}),
		Ctx: ctx,
	}
	// Terminate environment and flags.
	config.TermEnvFlags()
	// Terminate storage metrics.
	handlers.SetStorage(s)
	// Load metrics from file.
	if config.ArgsM.StoreFile != "" {
		err := helpers.LoadFromFile(&wg, logger, config.ArgsM)
		if err != nil {
			logger.Error("Error load from file: ", zap.Error(err))
		}
	}
	// Connect to database if DBURL exist.
	if config.ArgsM.DBURL != "" {
		err := handlers.StorageM.InitDB(config.ArgsM.DBURL)
		if err != nil {
			logger.Error("Connection to postrgres faild: ", zap.Error(err))
		}
		// Restore from database.
		if !config.ArgsM.Restore && config.ArgsM.DBURL != "" {
			err = handlers.StorageM.LoadMetricsDB()
			if err != nil {
				logger.Error("Error load metrics to database LoadMetricsDB: ", zap.Error(err))
			}
		}
	}
	// Add count wait group.
	wg.Add(1)
	// Wait signal from operation system.
	go helpers.WaitSignals(cancel, logger, &wg)
	// Add count wait group.
	wg.Add(1)
	// Sync metrics with file.
	go helpers.SyncFile(ctx, &wg, logger, config.ArgsM)
	// Init chi router.
	r := chi.NewRouter()
	r.Route("/", handlers.Router)

	fmt.Printf("Build version:%s \n", buildVersion)
	fmt.Printf("Build date:%s \n", buildDate)
	fmt.Printf("Build commit:%s \n", buildCommit)

	// Start http server.
	go func() {
		err := http.ListenAndServe(config.ArgsM.Address, r)
		if err != nil {
			logger.Error("Error http server CHI: ", zap.Error(err))
		}
	}()
	// Add count wait group.
	wg.Wait()
}
