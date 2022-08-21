package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/server/handlers"
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

// Init wait group.
var wg sync.WaitGroup

func main() {
	// Default context and cancel.
	ctx, cancel := context.WithCancel(context.Background())
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
		err := loadFromFile(config.ArgsM)
		if err != nil {
			logrus.Error("Error load from file: ", err)
		}
	}
	// Connect to database if DBURL exist.
	if config.ArgsM.DBURL != "" {
		err := handlers.StorageM.InitDB(config.ArgsM.DBURL)
		if err != nil {
			logrus.Error("Connection to postrgres faild: ", err)
		}
		// Restore from database.
		if !config.ArgsM.Restore && config.ArgsM.DBURL != "" {
			err = handlers.StorageM.LoadMetricsDB()
			if err != nil {
				logrus.Error("Error load metrics to database LoadMetricsDB: ", err)
			}
		}
	}
	// Add count wait group.
	wg.Add(1)
	// Wait signal from operation system.
	go waitSignals(cancel)
	// Add count wait group.
	wg.Add(1)
	// Sync metrics with file.
	go syncFile(config.ArgsM, ctx)
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
			logrus.Error("Error http server CHI: ", err)
		}
	}()
	// Add count wait group.
	wg.Wait()
}

// Load metrics from file storage.
func loadFromFile(env config.Args) error {
	if env.Restore && fileExist(env.StoreFile) {
		file, err := os.ReadFile(env.StoreFile)
		if err != nil {
			logrus.Error("Error open file for writing: ", err)
			wg.Done()
			return err
		}
		handlers.StorageM.LoadMetricsFile(file)
	}
	return nil
}

// Sync metrics with file storage.
func syncFile(env config.Args, ctx context.Context) {
	if env.StoreFile == "" {
		for {
			<-ctx.Done()
			logrus.Info("File syncing is down")
			wg.Done()
			return
		}
	} else {
		if env.StoreInterval == 0 {
			metrics, err := handlers.StorageM.GetMetricsJSON()
			if err != nil {
				logrus.Error("Error getting metrics format JSON GetMetricsJSON: ", err)
			}
			file, err := os.Create(env.StoreFile)
			if err != nil {
				logrus.Error("Error open file for writing: ", err)
			}
			defer file.Close()

			data, err := json.Marshal(metrics)
			if err != nil {
				logrus.Error("Error marshaling metrics : ", err)
			}
			_, err = file.Write(data)
			if err != nil {
				logrus.Error("Error write data to file: ", err)
			}
			for {
				<-ctx.Done()
				logrus.Info("File syncing is down")
				wg.Done()
				return
			}
		} else {
			for {
				select {
				case <-ctx.Done():
					logrus.Info("File syncing is down")
					wg.Done()
					return
				case <-time.After(env.StoreInterval):
					metrics, _ := handlers.StorageM.GetMetricsJSON()
					file, err := os.OpenFile(env.StoreFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
					if err != nil {
						logrus.Error("Error open file for writing: ", err)
					}
					defer file.Close()

					data, err := json.Marshal(metrics)
					if err != nil {
						logrus.Error("Error marshaling metrics : ", err)
					}
					_, err = file.Write(data)
					if err != nil {
						logrus.Error("Error writing data to file: ", err)
					}
				}
			}
		}
	}
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
func waitSignals(cancel context.CancelFunc) {
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
			logrus.Info("Terminate signal OS!")
			return
		}
	}
}
