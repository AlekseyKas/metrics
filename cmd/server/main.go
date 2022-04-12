package main

import (
	"log"
	"net/http"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/caarlos0/env"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
)

type Param struct {
	ADDRESS string `env:"ADDRESS"`
}

func main() {

	// var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: structs.Map(storage.Metrics{}),
	}
	handlers.SetStorage(s)
	a := GetParam()
	r := chi.NewRouter()
	r.Route("/", handlers.Router)
	http.ListenAndServe(a.ADDRESS, r)
}

//get param from env
func GetParam() Param {
	var param Param
	err := env.Parse(&param)
	if err != nil {
		log.Fatal(err)
	}
	if param.ADDRESS == "" {
		param.ADDRESS = "127.0.0.1:8080"
	}
	return param
}
