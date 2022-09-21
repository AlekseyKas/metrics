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
	"github.com/AlekseyKas/metrics/internal/server/grpc"
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
	// Terminate environment and flags.
	config.TermEnvFlags()
	// Send logger to storage
	storage.InitLogger(logger)
	// Inint storage server.
	s := &storage.MetricsStore{
		MM:  structs.Map(storage.Metrics{}),
		Ctx: ctx,
	}
	// Load metrics from file.
	if config.ArgsM.StoreFile != "" {
		err = helpers.LoadFromFile(logger, config.ArgsM)
		if err != nil {
			logger.Error("Error load from file: ", zap.Error(err))
		}
	}
	// Start http server
	if !config.ArgsM.GRPC {
		err = runHTTPserver(ctx, cancel, logger, &wg, s)
		if err != nil {
			logger.Error("Error run http server: ", zap.Error(err))
		}
	} else {
		runGRPCserver(ctx, cancel, logger, &wg, s)
	}
}

func runGRPCserver(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger, wg *sync.WaitGroup, s *storage.MetricsStore) {
	var err error
	// Connect to database if DBURL exist.
	if config.ArgsM.DBURL != "" {
		err = grpc.GRPCSrv.StorageM.InitDB(config.ArgsM.DBURL)
		if err != nil {
			logger.Error("Connection to postrgres faild: ", zap.Error(err))
		}
		// Restore from database.
		if !config.ArgsM.Restore && config.ArgsM.DBURL != "" {
			err = grpc.GRPCSrv.StorageM.LoadMetricsDB()
			if err != nil {
				logger.Error("Error load metrics to database LoadMetricsDB: ", zap.Error(err))
			}
		}
	}

	// Add count wait group.
	wg.Add(1)
	// Sync metrics with file.
	go helpers.SyncFile(ctx, wg, logger, config.ArgsM)
	fmt.Printf("Build version:%s \n", buildVersion)
	fmt.Printf("Build date:%s \n", buildDate)
	fmt.Printf("Build commit:%s \n", buildCommit)
	// Add count wait group.
	wg.Add(1)
	// Wait signal from operation system.
	go helpers.WaitSignals(cancel, logger, wg, nil)

	// Start gRPC server.
	go func() {
		srv := grpc.New(logger, s, config.ArgsM)
		srv.Start()
	}()
	// Add count wait group.
	wg.Wait()
}

func runHTTPserver(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger, wg *sync.WaitGroup, s *storage.MetricsStore) error {
	var err error
	// Send logger to handlers
	handlers.InitLogger(logger)
	// Init config
	handlers.InitConfig(config.ArgsM)
	// Terminate storage metrics.
	handlers.SetStorage(s)

	// Connect to database if DBURL exist.
	if config.ArgsM.DBURL != "" {
		err = handlers.StorageM.InitDB(config.ArgsM.DBURL)
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
	// Sync metrics with file.
	go helpers.SyncFile(ctx, wg, logger, config.ArgsM)
	// Init chi router.
	r := chi.NewRouter()
	r.Route("/", handlers.Router)

	fmt.Printf("Build version:%s \n", buildVersion)
	fmt.Printf("Build date:%s \n", buildDate)
	fmt.Printf("Build commit:%s \n", buildCommit)
	// Init http server
	var srv = http.Server{Addr: config.ArgsM.Address, Handler: r}
	// Add count wait group.
	wg.Add(1)
	// Wait signal from operation system.
	go helpers.WaitSignals(cancel, logger, wg, &srv)
	// Start http server.
	go func() {
		switch err = srv.ListenAndServe(); err {
		case nil:
		case http.ErrServerClosed:
		default:
			logger.Error("Error http server CHI: ", zap.Error(err))
		}
	}()
	// Add count wait group.
	wg.Wait()
	if err != http.ErrServerClosed {
		return err
	} else {
		return nil
	}
}
