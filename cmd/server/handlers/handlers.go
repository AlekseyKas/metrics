package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/pprof" // подключаем пакет pprof
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"

	"github.com/AlekseyKas/metrics/cmd/server/database"
	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
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
	r.Use(CompressGzip)
	r.Use(DecompressGzip)
	r.Get("/", getMetrics())
	r.Get("/value/{typeMet}/{nameMet}", getMetric())
	r.Get("/ping", checkConnection())
	r.Post("/update/{typeMet}/{nameMet}/{value}", saveMetrics())
	r.Post("/update/", saveMetricsJSON())
	r.Post("/updates/", saveMetricsSlice())
	r.Post("/value/", getMetricsJSON())

	r.Get("/debug/pprof/", pprof.Index)
	r.Get("/debug/pprof/cmdline", pprof.Cmdline)
	r.Get("/debug/pprof/profile", pprof.Profile)
	r.Get("/debug/pprof/symbol", pprof.Symbol)
	r.Get("/debug/pprof/trace", pprof.Trace)
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

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		next.ServeHTTP(gzipBodyWriter{
			ResponseWriter: w,
			writer:         gz,
		}, r)
	})
}

func saveMetricsSlice() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		s := StorageM.GetSliceStruct()

		out, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logrus.Error("Error read body: ", err)
		}

		err = json.Unmarshal(out, &s)
		if err != nil {
			logrus.Error("Error unmarshaling request: ", err)
		}
		var typeMet string
		var nameMet string
		for i := 0; i < len(s); i++ {
			typeMet = s[i].MType
			nameMet = s[i].ID
			metrics := StorageM.GetMetrics()

			if config.ArgsM.Key != "" {
				b, err := compareHash(&s[i], []byte(config.ArgsM.Key))
				if err != nil {
					logrus.Error("Error compare hash of metrics: ", err)
				}
				if b {

					//update gauge
					if typeMet == "gauge" {
						if s[i].Value != nil {
							if metrics[nameMet] != gauge(*s[i].Value) {
								StorageM.ChangeMetric(nameMet, gauge(*s[i].Value), config.ArgsM)
								StorageM.ChangeMetricDB(nameMet, *s[i].Value, typeMet, config.ArgsM)
								rw.WriteHeader(http.StatusOK)
							}
						}
					}
					//update counter
					if typeMet == "counter" {
						var valueMetInt int
						if s[i].Delta != nil {
							if _, ok := metrics[nameMet]; ok {
								ii, err := strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
								if err != nil {
									rw.WriteHeader(http.StatusBadRequest)
									return
								}
								valueMetInt = int(*s[i].Delta) + ii
								StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
								StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
								rw.WriteHeader(http.StatusOK)
							} else {
								valueMetInt = int(*s[i].Delta)
								StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
								StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
								rw.WriteHeader(http.StatusOK)
							}
						}
					}
				}

			} else {
				if typeMet == "gauge" {
					if s[i].Value != nil {
						if metrics[nameMet] != gauge(*s[i].Value) {
							StorageM.ChangeMetric(nameMet, gauge(*s[i].Value), config.ArgsM)
							StorageM.ChangeMetricDB(nameMet, *s[i].Value, typeMet, config.ArgsM)
							rw.WriteHeader(http.StatusOK)
						}
					}
				}
				//update counter
				if typeMet == "counter" {
					var valueMetInt int
					if s[i].Delta != nil {
						if _, ok := metrics[nameMet]; ok {
							ii, err := strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
							if err != nil {
								rw.WriteHeader(http.StatusBadRequest)
								return
							}
							valueMetInt = int(*s[i].Delta) + ii
							StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
							StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
							rw.WriteHeader(http.StatusOK)
						} else {
							valueMetInt = int(*s[i].Delta)
							StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
							StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
							rw.WriteHeader(http.StatusOK)
						}
					}
				}
			}
		}
	})
}

func getMetricsJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		rw.Header().Add("Content-Type", "application/json")
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
					if config.ArgsM.Key != "" {
						calculateHash(&s, []byte(config.ArgsM.Key))
						if err != nil {
							logrus.Error("Error calculate hash: ", err)
						}
					}
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
					if config.ArgsM.Key != "" {
						calculateHash(&s, []byte(config.ArgsM.Key))
						if err != nil {
							logrus.Error("Error calculate hash: ", err)
						}
					}
					encoder := json.NewEncoder(&buf)
					err = encoder.Encode(s)
					if err != nil {
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
	}
}

func calculateHash(s *storage.JSONMetrics, key []byte) {
	var h hash.Hash
	// var err error
	switch s.MType {
	case "counter":
		data := (fmt.Sprintf("%s:counter:%d", s.ID, *s.Delta))
		h = hmac.New(sha256.New, key)
		h.Write([]byte(data))
		s.Hash = fmt.Sprintf("%x", h.Sum(nil))
	case "gauge":
		data := (fmt.Sprintf("%s:gauge:%f", s.ID, *s.Value))
		h = hmac.New(sha256.New, key)
		h.Write([]byte(data))
		s.Hash = fmt.Sprintf("%x", h.Sum(nil))
	}
	// return err
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
		metrics := StorageM.GetMetrics()
		typeMet := s.MType
		nameMet := s.ID
		if config.ArgsM.Key != "" {
			b, err := compareHash(&s, []byte(config.ArgsM.Key))
			if err != nil {
				logrus.Error("Error compare hash of metrics: ", err)
			}
			if !b {
				rw.WriteHeader(http.StatusBadRequest)
			}
		}

		if typeMet != "gauge" && typeMet != "counter" {
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}
		//update gauge
		if typeMet == "gauge" {
			if s.Value == nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				if metrics[nameMet] != gauge(*s.Value) {
					StorageM.ChangeMetric(nameMet, gauge(*s.Value), config.ArgsM)
					StorageM.ChangeMetricDB(nameMet, *s.Value, typeMet, config.ArgsM)
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
					StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
					StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
					rw.WriteHeader(http.StatusOK)
					return
				} else {
					valueMetInt = int(*s.Delta)
					StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
					StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
					rw.WriteHeader(http.StatusOK)
					return
				}
			} else {
				rw.WriteHeader(http.StatusBadRequest)
			}
		}
	}
}

func compareHash(s *storage.JSONMetrics, key []byte) (b bool, err error) {
	var h hash.Hash

	switch s.MType {
	case "counter":
		data := (fmt.Sprintf("%s:counter:%d", s.ID, *s.Delta))
		h = hmac.New(sha256.New, key)
		h.Write([]byte(data))
	case "gauge":
		data := (fmt.Sprintf("%s:gauge:%f", s.ID, *s.Value))
		h = hmac.New(sha256.New, key)
		h.Write([]byte(data))
	}
	h.Sum(nil)
	if fmt.Sprintf("%x", h.Sum(nil)) == s.Hash {
		b = true
	}
	return b, nil
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
			return
		}
		rw.Header().Add("Content-Type", "text/html")
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
		//typeMertic to URL
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
				return
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

func checkConnection() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if config.ArgsM.DBURL != "" {
			err := database.Conn.Ping(context.Background())
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
			} else {
				rw.WriteHeader(http.StatusOK)
			}
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}
