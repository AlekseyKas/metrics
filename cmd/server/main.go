package main

import (
	"context"
	"encoding/json"
	"flag"
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

	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	termEnvFlags()
	fmt.Println("")
	handlers.SetStorage(s)

	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go syncFile(config.ArgsM, ctx)
	wg.Add(1)
	go waitSignals(cancel)

	r := chi.NewRouter()
	r.Route("/", handlers.Router)
	go http.ListenAndServe(config.ArgsM.Address, r)

	wg.Wait()
}

func syncFile(env config.Args, ctx context.Context) {
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

		fmt.Println("22222222222222222222222222222222222222", config.ArgsM.Address, config.ArgsM.Restore, config.ArgsM.StoreFile, config.ArgsM.StoreInterval)
		if env.Restore && fileExist(env.StoreFile) {

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

func termEnvFlags() {
	// kong.Parse(&config.FlagsServer)
	flag.StringVar(&config.FlagsServer.Address, "a", "127.0.0.1:8080", "Address")
	flag.StringVar(&config.FlagsServer.StoreFIle, "f", "/tmp/devops-metrics-db.json", "File path store")
	flag.BoolVar(&config.FlagsServer.Restore, "r", true, "Restire drom file")
	flag.DurationVar(&config.FlagsServer.StoreInterval, "i", 300000000000, "Interval store file")
	flag.Parse()
	fmt.Println(config.FlagsServer)
	env := config.LoadConfig()
	envADDR, _ := os.LookupEnv("ADDRESS")
	if envADDR == "" {
		fmt.Println(":;;;;;;;;;;;")
		config.ArgsM.Address = config.FlagsServer.Address
	} else {
		fmt.Println("sdsdsdsd")

		config.ArgsM.Address = env.Address
	}
	envRest, _ := os.LookupEnv("RESTORE")
	if envRest == "" {
		config.ArgsM.Restore = config.FlagsServer.Restore
	} else {
		config.ArgsM.Restore = env.Restore
	}
	envStoreint, _ := os.LookupEnv("STORE_INTERVAL")
	if envStoreint == "" {
		config.ArgsM.StoreInterval = config.FlagsServer.StoreInterval
	} else {
		config.ArgsM.StoreInterval = env.StoreInterval
	}
	envFile, _ := os.LookupEnv("STORE_FILE")
	if envFile == "" {
		config.ArgsM.StoreFile = config.FlagsServer.StoreFIle
	} else {
		config.ArgsM.StoreFile = env.StoreFile
	}
	fmt.Println("----------------", "Env address: ", env.Address, "Env Pollinterval: ", env.PollInterval, "Env ReportInterval: ", env.ReportInterval, "Env Restore: ", env.Restore, "Env Storefile: ", env.StoreFile, "Env Storeinsterval: ", env.StoreInterval)
	fmt.Println("Env address: ", env.Address, "Env Pollinterval: ", env.PollInterval, "Env ReportInterval: ", env.ReportInterval, "Env Restore: ", env.Restore, "Env Storefile: ", env.StoreFile, "Env Storeinsterval: ", env.StoreInterval, "----------------")
	fmt.Println("==============", "Flag address: ", config.ArgsM.Address, "Flag Pollinterval: ", config.ArgsM.PollInterval, "Flag ReportInterval: ", config.ArgsM.ReportInterval, "Flag Restore: ", config.ArgsM.Restore, "Flag Storefile: ", config.ArgsM.StoreFile, "Flag Storeinsterval: ", config.ArgsM.StoreInterval, "===================")

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
