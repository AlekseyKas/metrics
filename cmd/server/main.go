package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/caarlos0/env"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

type Param struct {
	Address       string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreInterval int    `env:"STORE_INTERVAL" envDefault:"5"`
	StoreFile     string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool   `env:"RESTORE" envDefault:"true"`
}

var wg sync.WaitGroup

func main() {

	// var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	handlers.SetStorage(s)
	env := GetParam()

	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go syncFile(env, ctx)
	wg.Add(1)
	go waitSignals(cancel)

	r := chi.NewRouter()
	r.Route("/", handlers.Router)
	go http.ListenAndServe(env.Address, r)

	wg.Wait()
}

//get param from env
func GetParam() Param {
	var Parametrs Param
	err := env.Parse(&Parametrs)
	if err != nil {
		log.Fatal(err)
	}
	// if Parametrs.Address == "" {
	// 	Parametrs.Address = "127.0.0.1:8080"
	// }
	// if Parametrs.StoreInterval == 0 {
	// 	Parametrs.StoreInterval = 5
	// }

	// f, bo := os.LookupEnv("STORE_FILE")
	// if f == "" && bo == true {
	// 	Parametrs.StoreFile = "none"
	// }
	// if f == "" && bo == false {
	// 	Parametrs.StoreFile = "/tmp/devops-metrics-db.json"
	// }
	// s, _ := os.LookupEnv("RESTORE")
	// switch s {
	// case "":
	// 	Parametrs.Restore = true
	// case "true":
	// 	Parametrs.Restore = true
	// case "false":
	// 	Parametrs.Restore = false
	// }
	return Parametrs
}

func syncFile(env Param, ctx context.Context) {
	if env.StoreFile == "" {
		for {
			select {
			case <-ctx.Done():
				logrus.Info("File syncing is down")
				wg.Done()
				return
			case <-time.After(time.Duration(env.StoreInterval) * time.Second):
				logrus.Info("Writing to disk disable.")
			}
		}
	} else {
		//restore data from file

		if env.Restore && fileExist(env.StoreFile) {
			file, err := os.ReadFile(env.StoreFile)
			if err != nil {
				logrus.Error("Error open file for writing: ", err)
				wg.Done()
				return
			}
			handlers.StorageM.LoadMetricsFile(file)
		}
		for {
			select {
			case <-ctx.Done():
				logrus.Info("File syncing is down")
				wg.Done()
				return
			case <-time.After(time.Duration(env.StoreInterval) * time.Second):
				metrics := handlers.StorageM.GetMetrics()
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
			logrus.Info("File syncing is terminate!")
			cancel()
			wg.Done()
			return
		}
	}
}
