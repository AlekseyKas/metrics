package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/caarlos0/env"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
)

type Param struct {
	Address string `env:"ADDRESS"`
	// StoreInterval int    `env: "STORE_INTERVAL"`
	// StoreFile     string `env: "STORE_FILE"`
	// Restore       bool   `env: "RESTORE"`
}

func main() {

	// var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	handlers.SetStorage(s)
	env := GetParam()
	fmt.Println(env)
	r := chi.NewRouter()
	r.Route("/", handlers.Router)
	http.ListenAndServe(env.Address, r)
}

//get param from env
func GetParam() Param {
	var param Param

	err := env.Parse(&param)
	if err != nil {
		log.Fatal(err)
	}
	if param.Address == "" {
		param.Address = "127.0.0.1:8080"
	}
	// if param.StoreInterval == 0 {
	// 	param.StoreInterval = 300
	// }
	// if param.StoreFile == "" {
	// 	param.StoreFile = "/tmp/devops-metrics-db.json"
	// }
	// if param.Restore == false {

	// }
	return param
}
