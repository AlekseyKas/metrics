package main

import (
	"net/http"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-chi/chi"
)

func main() {

	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	handlers.SetStorage(s)

	r := chi.NewRouter()
	r.Route("/", handlers.Router)
	http.ListenAndServe(":8080", r)
}
