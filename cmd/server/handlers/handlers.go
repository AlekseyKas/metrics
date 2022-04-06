package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

//init typs
type gauge float64
type counter int64

var storageM storage.Storage

func SetStorage(s storage.Storage) {
	storageM = s
}

//router
func Router(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", getMetrics())
	r.Get("/value/{typeMet}/{nameMet}", getMetric())
	r.Post("/update/{typeMet}/{nameMet}/{value}", saveMetrics())
	// r.Post("/update/{typeMet}/{nameMet}/{value}", saveMetrics())

}

//Get all metrics
func getMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		metrics := storageM.GetMetrics()
		jsonMetrics, err := json.Marshal(metrics)
		if err != nil {
			logrus.Error(err)
		}
		rw.Write(jsonMetrics)
		rw.Header().Add("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
	}
}

//Get value metric
func getMetric() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		metrics := storageM.GetMetrics()

		typeMet := chi.URLParam(req, "typeMet")
		nameMet := chi.URLParam(req, "nameMet")

		switch typeMet {
		case "gauge":
			typeMet = "gauge"
		case "counter":
			typeMet = "counter"
		}

		if typeMet != "gauge" && typeMet != "counter" {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if typeMet == "gauge" && nameMet == "PollCount" {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		if _, ok := metrics[nameMet]; ok {
			if typeMet == "counter" {
				rw.Write([]byte(fmt.Sprintf("%v", metrics[nameMet])))
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusOK)
				return
			}
		} else {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		}

		if typeMet == "gauge" && nameMet != "PollCount" {
			rw.Write([]byte(fmt.Sprintf("%v", metrics[nameMet])))
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusOK)
			return
		}
	}
}

func saveMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		typeMet := chi.URLParam(req, "typeMet")
		nameMet := chi.URLParam(req, "nameMet")
		value := chi.URLParam(req, "value")

		metrics := storageM.GetMetrics()
		fmt.Println("############", metrics)

		//typeMertic to url
		switch typeMet {
		case "gauge":
			typeMet = "gauge"
		case "counter":
			typeMet = "counter"
		}
		if typeMet != "gauge" && typeMet != "counter" {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotImplemented)
		}
		//update gauge
		if typeMet == "gauge" && nameMet != "PollCount" {
			valueMetFloat, err := strconv.ParseFloat(value, 64)
			if err != nil {
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusBadRequest)
			}
			if err == nil {
				if metrics[nameMet] != gauge(valueMetFloat) {
					/////////////
					storageM.ChangeGauge(nameMet, gauge(valueMetFloat))
					// metrics[nameMet] = gauge(valueMetFloat)
					rw.Header().Add("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusOK)
				}
			}
		}
		//update counter
		if typeMet == "counter" {

			valueMetInt, err := strconv.Atoi(value)
			if err != nil {
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusBadRequest)
			}
			if err == nil {
				if _, ok := metrics[nameMet]; ok {
					i, err := strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
					if err != nil {
						rw.Header().Add("Content-Type", "text/plain")
						rw.WriteHeader(http.StatusBadRequest)
					}

					valueMetInt = valueMetInt + i
					// metrics[nameMet] = counter(valueMetInt)
					storageM.ChangeGauge(nameMet, counter(valueMetInt))

					////////////
					rw.Header().Add("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusOK)
				} else {
					storageM.ChangeGauge(nameMet, counter(valueMetInt))

					// metrics[nameMet] = counter(valueMetInt)
					//////////////
				}
			}
		}
	}
}
