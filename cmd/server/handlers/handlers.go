package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

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
	r.Post("/update/", saveMetricsJSON())
	r.Post("/value/", getMetricsJSON())

}

func getMetricsJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		rw.Header().Add("Content-Type", "application/json")

		// out, err := ioutil.ReadAll(req.Body)
		// if err != nil {
		// 	http.Error(rw, err.Error(), 500)
		// 	return
		// }
		// fmt.Println("uuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuot", string(out))
		s := storageM.GetStructJSON()
		// err = json.Unmarshal(out, &s)
		// if err != nil {
		// 	logrus.Error("Error unmarshaling request: ", err)
		// 	http.Error(rw, err.Error(), http.StatusInternalServerError)

		// }
		err := json.NewDecoder(req.Body).Decode(&s)
		if err != nil {
			logrus.Info(err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		metrics := storageM.GetMetrics()
		typeMet := s.MType
		nameMet := s.ID

		if typeMet != "gauge" && typeMet != "counter" {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if typeMet == "gauge" && nameMet == "PollCount" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		if _, ok := metrics[nameMet]; ok {
			if typeMet == "counter" && strings.Split(reflect.ValueOf(metrics[nameMet]).Type().String(), ".")[1] == "counter" {
				i, err := strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
				}
				a := int64(i)
				s.Delta = &a
				// toSend, err := json.Marshal(s)
				// if err != nil {
				// 	logrus.Error("Error marshaling struct to sending", err)
				// 	http.Error(rw, err.Error(), http.StatusInternalServerError)
				// }
				var buf bytes.Buffer
				// buf =
				encoder := json.NewEncoder(&buf)
				encoder.Encode(s)
				// logrus.Infof("%+v", string(toSend))
				// json.Encoder()
				// op := storageM.GetStructJSON()
				// json.Unmarshal(toSend, &op)

				// fmt.Println("9090909090", op)
				rw.Write(buf.Bytes())
				rw.WriteHeader(http.StatusOK)
				return
			}
		} else {
			rw.WriteHeader(http.StatusNotFound)
		}

		if typeMet == "gauge" && nameMet != "PollCount" {
			float, err := strconv.ParseFloat(fmt.Sprintf("%v", metrics[nameMet]), 64)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
			}
			s.Value = &float
			toSend, err := json.Marshal(s)
			if err != nil {
				logrus.Error("Error marshaling struct to sending", err)
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
			rw.Write([]byte(toSend))
			rw.WriteHeader(http.StatusOK)
			return
		}

	}
}

//save metrics
func saveMetricsJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		// out, err := ioutil.ReadAll(req.Body)
		// if err != nil {
		// 	rw.WriteHeader(http.StatusBadRequest)
		// 	return
		// }
		s := storageM.GetStructJSON()
		// err = json.Unmarshal(out, &s)
		// if err != nil {
		// 	logrus.Error("Error unmarshaling request: ", err)
		// 	rw.WriteHeader(http.StatusBadRequest)

		// }
		err := json.NewDecoder(req.Body).Decode(&s)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}

		metrics := storageM.GetMetrics()
		typeMet := s.MType
		nameMet := s.ID

		rw.Header().Add("Content-Type", "application/json")

		if typeMet != "gauge" && typeMet != "counter" {
			rw.WriteHeader(http.StatusNotImplemented)
		}
		//update gauge
		if typeMet == "gauge" {
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
			}
			if err == nil {
				if s.Value == nil {
					rw.WriteHeader(http.StatusInternalServerError)
				} else {
					if metrics[nameMet] != gauge(*s.Value) {
						storageM.ChangeMetric(nameMet, gauge(*s.Value))
						rw.WriteHeader(http.StatusOK)
					}
				}

			}
		}
		//update counter
		if typeMet == "counter" {
			// fmt.Println("8888888888888888888888888888", nameMet, typeMet, *s.Delta)
			var valueMetInt int
			if err == nil {
				if _, ok := metrics[nameMet]; ok {
					i, err := strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
					if err != nil {
						rw.WriteHeader(http.StatusBadRequest)
					}
					if s.Delta == nil {
						rw.WriteHeader(http.StatusInternalServerError)
					} else {
						valueMetInt = int(*s.Delta) + i
						// fmt.Println("60000000000000", i)
						// fmt.Println("66666666666666666666666", *s.Delta, nameMet, valueMetInt)
						storageM.ChangeMetric(nameMet, counter(valueMetInt))
						rw.WriteHeader(http.StatusOK)
					}
				} else {
					valueMetInt = int(*s.Delta)
					// fmt.Println("7777777777777777777777", *s.Delta, nameMet, valueMetInt)
					storageM.ChangeMetric(nameMet, counter(valueMetInt))
					rw.WriteHeader(http.StatusOK)
				}
			}
		}
	}
}

//Get all metrics
func getMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		metrics := storageM.GetMetrics()
		JSONMetrics, err := json.Marshal(metrics)
		if err != nil {
			logrus.Error(err)
		}

		rw.Write(JSONMetrics)
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
					storageM.ChangeMetric(nameMet, gauge(valueMetFloat))
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

					storageM.ChangeMetric(nameMet, counter(valueMetInt))

					rw.Header().Add("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusOK)
				} else {
					storageM.ChangeMetric(nameMet, counter(valueMetInt))
				}
			}
		}
	}
}
