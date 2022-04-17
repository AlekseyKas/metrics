package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

//init typs
type gauge float64
type counter int64

var StorageM storage.Storage

func SetStorage(s storage.Storage) {
	StorageM = s
}

//router
func Router(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(CompressGzip)
	r.Use(DecompressGzip)
	r.Get("/", getMetrics())
	r.Get("/value/{typeMet}/{nameMet}", getMetric())
	r.Post("/update/{typeMet}/{nameMet}/{value}", saveMetrics())
	r.Post("/update/", saveMetricsJSON())
	r.Post("/value/", getMetricsJSON())

}

type gzipBodyWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (gz gzipBodyWriter) Write(b []byte) (int, error) {
	return gz.writer.Write(b)
}

func DecompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			logrus.Info("ppppppppppppwwwwwwwwwwwwwwwwwwwwwwwwwwwww", r.Header.Get("Content-Encoding"))
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		gz.Close()
		r.Body = gz
		next.ServeHTTP(w, r)

	})
}

func CompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer gz.Close()
		logrus.Info("0000000000000000", r.Body, "[[[[[[[[[[[[", r.Header)
		// w.Header().Set("Accept", t)
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		next.ServeHTTP(gzipBodyWriter{
			ResponseWriter: w,
			writer:         gz,
		}, r)
	})
}

func getMetricsJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		rw.Header().Add("Content-Type", "application/json")
		fmt.Println()
		out, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			return
		}

		s := StorageM.GetStructJSON()
		err = json.Unmarshal(out, &s)
		if err != nil {
			logrus.Error("Error unmarshaling request: ", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		// logrus.Infof("%+v", s)

		metrics := StorageM.GetMetrics()
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
		if typeMet == "counter" {

			if _, ok := metrics[nameMet]; ok {

				if typeMet == "counter" && strings.Split(reflect.ValueOf(metrics[nameMet]).Type().String(), ".")[1] == "counter" {
					i, err := strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
					if err != nil {
						rw.WriteHeader(http.StatusBadRequest)
					}
					a := int64(i)
					s.Delta = &a
					var buf bytes.Buffer
					encoder := json.NewEncoder(&buf)
					err = encoder.Encode(s)
					if err != nil {
						logrus.Info(err)
						http.Error(rw, err.Error(), http.StatusBadRequest)
					}
					rw.Write(buf.Bytes())
					rw.WriteHeader(http.StatusOK)
					return
				} else {
					rw.WriteHeader(http.StatusNotFound)
				}
			} else {
				rw.WriteHeader(http.StatusNotFound)
			}
		}

		if typeMet == "gauge" {
			if _, ok := metrics[nameMet]; ok {

				if strings.Split(reflect.ValueOf(metrics[nameMet]).Type().String(), ".")[1] == "gauge" {

					float, err := strconv.ParseFloat(fmt.Sprintf("%v", metrics[nameMet]), 64)
					if err != nil {
						rw.WriteHeader(http.StatusBadRequest)
					}
					s.Value = &float

					var buf bytes.Buffer
					encoder := json.NewEncoder(&buf)
					err = encoder.Encode(s)
					if err != nil {
						logrus.Info(err)
						http.Error(rw, err.Error(), http.StatusBadRequest)
					}
					rw.Write(buf.Bytes())
					rw.WriteHeader(http.StatusOK)
					return
				} else {
					rw.WriteHeader(http.StatusNotFound)
				}
			} else {
				rw.WriteHeader(http.StatusNotFound)
			}

		}

		// if typeMet == "gauge" && nameMet != "PollCount" {
		// 	float, err := strconv.ParseFloat(fmt.Sprintf("%v", metrics[nameMet]), 64)
		// 	if err != nil {
		// 		rw.WriteHeader(http.StatusBadRequest)
		// 	}
		// 	s.Value = &float

		// 	var buf bytes.Buffer
		// 	encoder := json.NewEncoder(&buf)
		// 	err = encoder.Encode(s)
		// 	if err != nil {
		// 		logrus.Info(err)
		// 		http.Error(rw, err.Error(), http.StatusBadRequest)
		// 	}
		// 	rw.Write(buf.Bytes())
		// 	rw.WriteHeader(http.StatusOK)
		// 	return
		// } else {
		// 	rw.WriteHeader(http.StatusNotFound)
		// }
	}
}

//save metrics
func saveMetricsJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		out, err := ioutil.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		s := StorageM.GetStructJSON()
		err = json.Unmarshal(out, &s)
		if err != nil {
			logrus.Error("Error unmarshaling request: ", err)
			rw.WriteHeader(http.StatusBadRequest)

		}
		// logrus.Info(s)
		metrics := StorageM.GetMetrics()
		typeMet := s.MType
		nameMet := s.ID

		if typeMet != "gauge" && typeMet != "counter" {
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}
		//update gauge
		if typeMet == "gauge" {
			if s.Value == nil {
				rw.WriteHeader(http.StatusInternalServerError)
				logrus.Info("999999999999999999999999999999", nameMet, typeMet, s.Value)
				return
			} else {
				if metrics[nameMet] != gauge(*s.Value) {
					StorageM.ChangeMetric(nameMet, gauge(*s.Value), config.ArgsM)

					rw.WriteHeader(http.StatusOK)
					return
				}
			}
		}
		//update counter
		if typeMet == "counter" {

			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			var valueMetInt int
			if s.Delta != nil {
				if _, ok := metrics[nameMet]; ok {
					i, err := strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
					if err != nil {
						rw.WriteHeader(http.StatusBadRequest)
						return
					}
					valueMetInt = int(*s.Delta) + i
					logrus.Info("88888888888888888888", nameMet, typeMet, s.Value)

					StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
					rw.WriteHeader(http.StatusOK)
					return
				} else {
					valueMetInt = int(*s.Delta)
					StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
					rw.WriteHeader(http.StatusOK)
					return
				}
			} else {
				rw.WriteHeader(http.StatusBadRequest)
			}
		}
	}
}

//Get all metrics
func getMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		metrics := StorageM.GetMetrics()
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		err := encoder.Encode(metrics)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		rw.Header().Add("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
		rw.Write(buf.Bytes())
	}
}

//Get value metric
func getMetric() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		metrics := StorageM.GetMetrics()

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
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(fmt.Sprintf("%v", metrics[nameMet])))
				return
			}
		} else {

			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		}

		if typeMet == "gauge" && nameMet != "PollCount" {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(fmt.Sprintf("%v", metrics[nameMet])))
			return
		}
	}
}

func saveMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		typeMet := chi.URLParam(req, "typeMet")
		nameMet := chi.URLParam(req, "nameMet")
		value := chi.URLParam(req, "value")

		metrics := StorageM.GetMetrics()
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
					StorageM.ChangeMetric(nameMet, gauge(valueMetFloat), config.ArgsM)
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

					StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)

					rw.Header().Add("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusOK)
				} else {
					StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
				}
			}
		}
	}
}
