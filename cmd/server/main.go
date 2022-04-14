package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

func main() {

	// var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	handlers.SetStorage(s)
	env := config.LoadConfig()
	fmt.Println("eeeeeeeeeeeeeeeeeeeeeeeeeeeeee", env)
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

func syncFile(env config.Param, ctx context.Context) {
	if env.StoreFile == "" {
		fmt.Println("11111111111111111111111111111111111111", env)

		for {
			<-ctx.Done()
			logrus.Info("File syncing is down")
			wg.Done()
			return
		}
	} else {
		//restore data from file

		if env.Restore && fileExist(env.StoreFile) {
			fmt.Println("22222222222222222222222222222222222222", env)

			file, err := os.ReadFile(env.StoreFile)
			if err != nil {
				logrus.Error("Error open file for writing: ", err)
				wg.Done()
				return
			}
			handlers.StorageM.LoadMetricsFile(file)
		}
		if env.StoreInterval == 0 {
			fmt.Println("3333333333333333333333333333333333333333333333", env)

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
					fmt.Println("44444444444444444444444444444444444444444444", env)

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
			logrus.Info("File syncing is terminate!")
			cancel()
			wg.Done()
			return
		}
	}
}
