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

	"github.com/AlekseyKas/metrics/cmd/server/database"
	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

func main() {
	// pool, err := pgxpool.Connect(context.Background(), DBURL)
	// if err != nil {
	// 	log.Fatalf("Unable to connection to database: %v\n", err)
	// }
	// defer pool.Close()
	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	termEnvFlags()
	handlers.SetStorage(s)

	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go syncFile(config.ArgsM, ctx)
	wg.Add(1)
	go waitSignals(cancel)
	logrus.Info(";;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;", config.ArgsM.DBURL, "sss", config.ArgsM)
	// msg = ";;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;sss{ localhost:37753 /tmp/OaUjPOc 0s 0s 5m0s /tmp/devops-metrics-db.json true}"
	//DB connection
	if config.ArgsM.DBURL != "" {

		err := database.DBConnect()
		if err != nil {
			logrus.Error("Connection to postrgres faild: ", err)
		}
	}
	r := chi.NewRouter()
	r.Route("/", handlers.Router)
	go http.ListenAndServe(config.ArgsM.Address, r)

	wg.Wait()
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

func termEnvFlags() {
	// kong.Parse(&config.FlagsServer)
	flag.StringVar(&config.FlagsServer.Address, "a", "127.0.0.1:8080", "Address")
	flag.StringVar(&config.FlagsServer.DBURL, "d", "", "Database URL")
	flag.StringVar(&config.FlagsServer.StoreFIle, "f", "", "File path store")
	flag.StringVar(&config.FlagsServer.Key, "k", "", "Secret key")
	flag.BoolVar(&config.FlagsServer.Restore, "r", true, "Restire drom file")
	flag.DurationVar(&config.FlagsServer.StoreInterval, "i", 300000000000, "Interval store file")
	flag.Parse()
	fmt.Println(config.FlagsServer)
	env := config.LoadConfig()
	envADDR, _ := os.LookupEnv("ADDRESS")
	if envADDR == "" {
		config.ArgsM.Address = config.FlagsServer.Address
	} else {
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
	envKey, _ := os.LookupEnv("KEY")
	if envKey == "" {
		config.ArgsM.Key = config.FlagsServer.Key
	} else {
		config.ArgsM.Key = env.Key
	}
	// if envFile == "" {
	// 	config.ArgsM.StoreFile = config.FlagsServer.StoreFIle
	// } else {
	// 	config.ArgsM.StoreFile = env.StoreFile
	// }

	envFile, _ := os.LookupEnv("STORE_FILE")
	envDBURL, _ := os.LookupEnv("DATABASE_DSN")
	if envDBURL == "" && config.FlagsServer.DBURL == "" {
		if envFile == "" && config.FlagsServer.StoreFIle == "" {
			config.ArgsM.DBURL = env.DBURL
			config.ArgsM.StoreFile = ""
			config.ArgsM.Restore = false
		} else {
			if envFile == "" {
				config.ArgsM.StoreFile = config.FlagsServer.StoreFIle
			} else {
				config.ArgsM.StoreFile = env.StoreFile
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
