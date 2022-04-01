package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	// "github.com/go-chi/chi/v5"
)

//init typs
type gauge float64
type counter int64

//Struct for metrics
type Metrics struct {
	Alloc         gauge
	BuckHashSys   gauge
	Frees         gauge
	GCCPUFraction gauge
	GCSys         gauge
	HeapAlloc     gauge
	HeapIdle      gauge
	HeapInuse     gauge
	HeapObjects   gauge
	HeapReleased  gauge
	HeapSys       gauge
	LastGC        gauge
	Lookups       gauge
	MCacheInuse   gauge
	MCacheSys     gauge
	MSpanInuse    gauge
	MSpanSys      gauge
	Mallocs       gauge
	NextGC        gauge
	NumForcedGC   gauge
	NumGC         gauge
	OtherSys      gauge
	PauseTotalNs  gauge
	StackInuse    gauge
	StackSys      gauge
	Sys           gauge
	TotalAlloc    gauge

	PollCount   counter
	RandomValue gauge
}

var metrics Metrics = Metrics{}
var mapMetrics map[string]interface{} = structs.Map(metrics)

func (metrics *Metrics) Router(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", getMetrics(mapMetrics))
	r.Get("/value/{typeMet}/{nameMet}", getMetric(mapMetrics))
	r.Post("/update/{typeMet}/{nameMet}/{value}", SaveMetrics(mapMetrics))
}

func main() {

	r := chi.NewRouter()
	r.Route("/", metrics.Router)
	http.ListenAndServe(":8080", r)

}

func getMetrics(metrics map[string]interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		jsonMetrics, err := json.Marshal(metrics)
		if err != nil {
			logrus.Error(err)
		}
		rw.Write(jsonMetrics)
		rw.Header().Add("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
	}
}

func getMetric(mapMetrics map[string]interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		typeMet := chi.URLParam(req, "typeMet")
		nameMet := chi.URLParam(req, "nameMet")

		//typeMertic to url
		switch typeMet {
		case "gauge":
			typeMet = "gauge"
		case "counter":
			typeMet = "counter"
		}
		//nameMertic to url
		if _, ok := mapMetrics[nameMet]; ok {
			nameMet = chi.URLParam(req, "nameMet")
		} else {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		}

		if typeMet == "gauge" && nameMet == "PollCount" {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		}
		if typeMet == "counter" && nameMet == "PollCount" {
			rw.Write([]byte(fmt.Sprintf("%v", mapMetrics[nameMet])))
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusOK)
		}
		if typeMet == "counter" && nameMet != "PollCount" {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		}
		if typeMet == "gauge" && nameMet != "PollCount" {
			rw.Write([]byte(fmt.Sprintf("%v", mapMetrics[nameMet])))
			fmt.Println("@@@@@@@@@@@@@@@@@", mapMetrics[nameMet])
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusOK)
		}
	}
}

func SaveMetrics(mapMetrics map[string]interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		typeMet := chi.URLParam(req, "typeMet")
		nameMet := chi.URLParam(req, "nameMet")
		value := chi.URLParam(req, "value")

		//typeMertic to url
		switch typeMet {
		case "gauge":
			typeMet = "gauge"
		case "counter":
			typeMet = "counter"
		}
		//nameMertic to url
		if _, ok := mapMetrics[nameMet]; ok {
			nameMet = chi.URLParam(req, "nameMet")
		} else {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		}
		//update gauge
		if typeMet == "gauge" && nameMet != "PollCount" {
			valueMetFloat, err := strconv.ParseFloat(value, 64)
			if err != nil {
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusNotFound)
				logrus.Error("Error parse float64: ", err)
			}
			if mapMetrics[nameMet] != gauge(valueMetFloat) {
				mapMetrics[nameMet] = gauge(valueMetFloat)
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusOK)
			}
		}
		//update counter
		if typeMet == "counter" && nameMet == "PollCount" {
			valueMetInt, _ := strconv.Atoi(value)
			if mapMetrics[nameMet] != valueMetInt {
				i, err := strconv.Atoi(fmt.Sprintf("%v", mapMetrics[nameMet]))
				if err != nil {
					rw.Header().Add("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusNotFound)
					logrus.Error("Error parse value to int", err)
				}
				valueMetInt = valueMetInt + i
				mapMetrics[nameMet] = counter(valueMetInt)
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusOK)
			}
		}

	}
}
