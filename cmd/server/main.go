package main

import (
	"context"
	"encoding/json"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"

	"github.com/AlekseyKas/metrics/cmd/server/database"
	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
)

var wg sync.WaitGroup

func main() {
	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}

	config.TermEnvFlags()
	handlers.SetStorage(s)
	logrus.Info(config.ArgsM.StoreFile)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go waitSignals(cancel)

	//load metrics from file
	if config.ArgsM.StoreFile != "" {
		err := loadFromFile(config.ArgsM)
		if err != nil {
			logrus.Error("Error load from file: ", err)
		}
	}
	//DB connection
	if config.ArgsM.DBURL != "" {
		err := database.DBConnect(config.ArgsM.DBURL)
		if err != nil {
			logrus.Error("Connection to postrgres faild: ", err)
		}
		jm, err := handlers.StorageM.GetMetricsJSON()
		if err != nil {
			logrus.Error("Error getting metricsJSON for database: ", err)
		}
		handlers.StorageM.InitDB(jm)
		//restore from DB
		if !config.ArgsM.Restore && config.ArgsM.DBURL != "" {
			handlers.StorageM.LoadMetricsDB()
		}
	}
	wg.Add(1)
	//sync metrics with file
	go syncFile(config.ArgsM, ctx)
	r := chi.NewRouter()
	r.Route("/", handlers.Router)
	go http.ListenAndServe(config.ArgsM.Address, r)

	wg.Wait()
}

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
			metrics, _ := handlers.StorageM.GetMetricsJSON()
			file, err := os.Create(env.StoreFile)
			if err != nil {
				logrus.Error("Error open file for writing: ", err)
			}
			defer file.Close()

			data, err := json.Marshal(metrics)
			if err != nil {
				logrus.Error("Error marshaling metrics : ", err)
			}
			file.Write(data)
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
					file.Write(data)
				}
			}
		}
	}
}

func fileExist(file string) bool {
	var b bool
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return b
	}
	b = true
	return b
}

//wating signals
func waitSignals(cancel context.CancelFunc) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		sig := <-terminate
		switch sig {
		case os.Interrupt:
			if config.ArgsM.DBURL != "" {
				database.DBClose()
			}
			cancel()
			wg.Done()
			logrus.Info("Terminate signal OS!")
			return
		}
	}
}
