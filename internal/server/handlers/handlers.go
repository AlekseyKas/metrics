package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/crypto"
	"github.com/AlekseyKas/metrics/internal/storage"
)

// Init type metrics gauge and counter
type gauge float64
type counter int64

// Init server storage
var StorageM storage.Storage

// Terminate storage server
func SetStorage(s storage.Storage) {
	StorageM = s
}

// Init logger.
var Logger *zap.Logger

// Set logger.
func InitLogger(logger *zap.Logger) {
	Logger = logger
}

// Init Args
var Args config.Args

// Set args
func InitConfig(args config.Args) {
	Args = args
}

// Handlers server
func Router(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(CheckSubnet)
	r.Use(Dencrypt)
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

func CheckSubnet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Args.TrustedSubnet != "" {
			ipStr := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStr)
			_, n, err := net.ParseCIDR(Args.TrustedSubnet)
			if err != nil {
				Logger.Error("Error parse parse trusted subnet: ", zap.Error(err))
			}
			result := n.Contains(ip)
			if result {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func Dencrypt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Args.PrivateKey != "" {

			out, err := io.ReadAll(r.Body)
			if err != nil {
				Logger.Error("Error read body: ", zap.Error(err))
			}
			data, err := crypto.DecryptData(out, Args.PrivateKey)
			if err != nil {
				Logger.Error("Error decrypt data: ", zap.Error(err))
			}
			r.ContentLength = int64(len(data))
			r.Body = io.NopCloser(bytes.NewBuffer(data))
			next.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
			return
		}
	})
}

// Init type for zipping
type gzipBodyWriter struct {
	http.ResponseWriter
	writer io.Writer
}

// Write and compress data
func (gz gzipBodyWriter) Write(b []byte) (int, error) {
	return gz.writer.Write(b)
}

// Decompress data
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

// Compress data
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

// Handler for saving metrics
func saveMetricsSlice() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		s := StorageM.GetSliceStruct()

		out, err := io.ReadAll(r.Body)
		if err != nil {
			Logger.Error("Error read body: ", zap.Error(err))
		}
		err = json.Unmarshal(out, &s)
		if err != nil {
			Logger.Error("Error unmarshaling request: ", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		var typeMet string
		var nameMet string
		for i := 0; i < len(s); i++ {
			typeMet = s[i].MType
			nameMet = s[i].ID
			metrics := StorageM.GetMetrics()

			if config.ArgsM.Key != "" {
				var b bool
				b, err = compareHash(&s[i], []byte(config.ArgsM.Key))
				if err != nil {
					Logger.Error("Error compare hash of metrics: ", zap.Error(err))
				}
				if b {

					//update gauge
					if typeMet == "gauge" {
						if s[i].Value != nil {
							if metrics[nameMet] != gauge(*s[i].Value) {
								err = StorageM.ChangeMetric(nameMet, gauge(*s[i].Value), config.ArgsM)
								if err != nil {
									Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
								}
								err = StorageM.ChangeMetricDB(nameMet, *s[i].Value, typeMet, config.ArgsM)
								if err != nil {
									Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
								}
								rw.WriteHeader(http.StatusOK)
							}
						}
					}
					//update counter
					if typeMet == "counter" {
						var valueMetInt int
						if s[i].Delta != nil {
							if _, ok := metrics[nameMet]; ok {
								var ii int
								ii, err = strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
								if err != nil {
									rw.WriteHeader(http.StatusBadRequest)
									return
								}
								valueMetInt = int(*s[i].Delta) + ii
								err = StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
								if err != nil {
									Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
								}
								err = StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
								if err != nil {
									Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
								}
								rw.WriteHeader(http.StatusOK)
							} else {
								valueMetInt = int(*s[i].Delta)
								err = StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
								if err != nil {
									Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
								}
								err = StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
								if err != nil {
									Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
								}
								rw.WriteHeader(http.StatusOK)
							}
						}
					}
				}

			} else {
				if typeMet == "gauge" {
					if s[i].Value != nil {
						if metrics[nameMet] != gauge(*s[i].Value) {
							err = StorageM.ChangeMetric(nameMet, gauge(*s[i].Value), config.ArgsM)
							if err != nil {
								Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
							}
							err = StorageM.ChangeMetricDB(nameMet, *s[i].Value, typeMet, config.ArgsM)
							if err != nil {
								Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
							}
							rw.WriteHeader(http.StatusOK)
						}
					}
				}
				//update counter
				if typeMet == "counter" {
					var valueMetInt int
					if s[i].Delta != nil {
						if _, ok := metrics[nameMet]; ok {
							var ii int
							ii, err = strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
							if err != nil {
								rw.WriteHeader(http.StatusBadRequest)
								return
							}
							valueMetInt = int(*s[i].Delta) + ii
							err = StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
							if err != nil {
								Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
							}
							err = StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
							if err != nil {
								Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
							}
							rw.WriteHeader(http.StatusOK)
						} else {
							valueMetInt = int(*s[i].Delta)
							err = StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
							if err != nil {
								Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
							}
							err = StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
							if err != nil {
								Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
							}
							rw.WriteHeader(http.StatusOK)
						}
					}
				}
			}
		}
	})
}

// Handler for getting metrics format JSON
func getMetricsJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		rw.Header().Add("Content-Type", "application/json")
		out, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			return
		}

		s := StorageM.GetStructJSON()
		err = json.Unmarshal(out, &s)
		if err != nil {
			Logger.Error("Error unmarshaling request: ", zap.Error(err))
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
							Logger.Error("Error calculate hash: ", zap.Error(err))
						}
					}
					encoder := json.NewEncoder(&buf)
					err = encoder.Encode(s)
					if err != nil {
						Logger.Error("Error encode metrics", zap.Error(err))
						http.Error(rw, err.Error(), http.StatusBadRequest)
					}
					_, err = rw.Write(buf.Bytes())
					if err != nil {
						Logger.Error("Error write data in buff: ", zap.Error(err))
					}
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
							Logger.Error("Error calculate hash: ", zap.Error(err))
						}
					}
					encoder := json.NewEncoder(&buf)
					err = encoder.Encode(s)
					if err != nil {
						http.Error(rw, err.Error(), http.StatusBadRequest)
					}
					_, err = rw.Write(buf.Bytes())
					if err != nil {
						Logger.Error("Error write bytes to req: ", zap.Error(err))
					}
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

// Calculate sha256
func calculateHash(s *storage.JSONMetrics, key []byte) {
	// Init hash
	var h hash.Hash
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
}

// Save metrics from format JSON
func saveMetricsJSON() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		out, err := io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		s := StorageM.GetStructJSON()
		err = json.Unmarshal(out, &s)
		if err != nil {
			Logger.Error("Error unmarshaling request: ", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		metrics := StorageM.GetMetrics()
		typeMet := s.MType
		nameMet := s.ID
		if config.ArgsM.Key != "" {
			var b bool
			b, err = compareHash(&s, []byte(config.ArgsM.Key))
			if err != nil {
				Logger.Error("Error compare hash of metrics: ", zap.Error(err))
			}
			if !b {
				rw.WriteHeader(http.StatusBadRequest)
			}
		}

		if typeMet != "gauge" && typeMet != "counter" {
			rw.WriteHeader(http.StatusNotImplemented)
			return
		}
		// Update gauge
		if typeMet == "gauge" {
			if s.Value == nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				if metrics[nameMet] != gauge(*s.Value) {
					err = StorageM.ChangeMetric(nameMet, gauge(*s.Value), config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					}
					err = StorageM.ChangeMetricDB(nameMet, *s.Value, typeMet, config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
					}
					rw.WriteHeader(http.StatusOK)
					return
				}
			}
		}
		// Update counter
		if typeMet == "counter" {
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			var valueMetInt int
			if s.Delta != nil {
				if _, ok := metrics[nameMet]; ok {
					var i int
					i, err = strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
					if err != nil {
						rw.WriteHeader(http.StatusBadRequest)
						return
					}
					valueMetInt = int(*s.Delta) + i
					err = StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					}
					err = StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
					}
					rw.WriteHeader(http.StatusOK)
					return
				} else {
					valueMetInt = int(*s.Delta)
					err = StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					}
					err = StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
					}
					rw.WriteHeader(http.StatusOK)
					return
				}
			} else {
				rw.WriteHeader(http.StatusBadRequest)
			}
		}
	}
}

// Compare hashe metrics
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
		_, err = h.Write([]byte(data))
		if err != nil {
			Logger.Error("Error write data hash: ", zap.Error(err))
		}
	}
	h.Sum(nil)
	if fmt.Sprintf("%x", h.Sum(nil)) == s.Hash {
		b = true
	}
	return b, nil
}

// Getting all metrics
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
		_, err = rw.Write(buf.Bytes())
		if err != nil {
			Logger.Error("Error write bytes to req: ", zap.Error(err))
		}
	}
}

// Get value metric
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
				_, err := rw.Write([]byte(fmt.Sprintf("%v", metrics[nameMet])))
				if err != nil {
					Logger.Error("Error write bytes to req: ", zap.Error(err))
				}
				return
			}
		} else {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		}
		if typeMet == "gauge" && nameMet != "PollCount" {
			rw.Header().Add("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusOK)
			_, err := rw.Write([]byte(fmt.Sprintf("%v", metrics[nameMet])))
			if err != nil {
				Logger.Error("Error write bytes to req: ", zap.Error(err))
			}
			return
		}
	}
}

func saveMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		typeMet := chi.URLParam(req, "typeMet")
		nameMet := chi.URLParam(req, "nameMet")
		value := chi.URLParam(req, "value")

		// Get metrics from memory
		metrics := StorageM.GetMetrics()

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
		// Update gauge
		if typeMet == "gauge" && nameMet != "PollCount" {
			valueMetFloat, err := strconv.ParseFloat(value, 64)
			if err != nil {
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusBadRequest)
				return
			} else {
				if metrics[nameMet] != gauge(valueMetFloat) {
					err = StorageM.ChangeMetric(nameMet, gauge(valueMetFloat), config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					}
					rw.Header().Add("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusOK)
				}
			}
		}
		// Update counter
		if typeMet == "counter" {
			valueMetInt, err := strconv.Atoi(value)
			if err != nil {
				rw.Header().Add("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusBadRequest)
			}
			if err == nil {
				if _, ok := metrics[nameMet]; ok {
					var i int
					i, err = strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
					if err != nil {
						rw.Header().Add("Content-Type", "text/plain")
						rw.WriteHeader(http.StatusBadRequest)
					}
					valueMetInt = valueMetInt + i
					err = StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					}
					rw.Header().Add("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusOK)
				} else {
					err = StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
					if err != nil {
						Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					}
				}
			}
		}
	}
}

// Checking connection to database
func checkConnection() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if config.ArgsM.DBURL != "" {

			err := StorageM.CheckConnection()
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
