package main

import (
	"net/http"

	"github.com/AlekseyKas/metrics/cmd/server/handlers"
	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	r.Route("/", handlers.Metric.Router)
	http.ListenAndServe(":8080", r)

}
