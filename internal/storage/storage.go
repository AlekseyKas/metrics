package storage

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
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

type JSONMetrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type MetricsStore struct {
	mux       sync.Mutex
	MM        map[string]interface{}
	PollCount int
}

type ValidStruct struct {
	ID    interface{} `json:"id"`              // имя метрики
	MType interface{} `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta interface{} `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value interface{} `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type StorageAgent interface {
	GetMetrics() map[string]interface{}
	ChangeMetrics(metrics runtime.MemStats) error

	GetMetricsJSON() ([]JSONMetrics, error)
}

type Storage interface {
	GetMetrics() map[string]interface{}
	ChangeMetric(nameMet string, value interface{}) error
	GetStructJSON() JSONMetrics
	ValidStruct(out []byte) bool
	// ChangeMetricJson(out []byte)
}

func (m *MetricsStore) ValidStruct(out []byte) bool {
	var b bool

	v := ValidStruct{}
	err := json.Unmarshal(out, &v)
	if err != nil {
		logrus.Error("Error unmarshaling in validation: ", err)
	}
	if v.MType == "counter" && v.Delta != nil {
		if reflect.ValueOf(v.Delta).Type().String() == "float64" {
			b = true
		}
	}
	if v.MType == "gauge" && v.Value != nil {
		if reflect.ValueOf(v.Value).Type().String() == "float64" {
			b = true
		}
	}
	fmt.Println(reflect.ValueOf(v.Delta).Type().String())
	return b
}

func (m *MetricsStore) GetStructJSON() JSONMetrics {
	s := JSONMetrics{}
	return s
}

func (m *MetricsStore) GetMetricsJSON() ([]JSONMetrics, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	// values := make(map[string]interface{}, (len(m.MM)))
	var j []JSONMetrics
	for k, v := range m.MM {
		if strings.Split(reflect.ValueOf(v).Type().String(), ".")[1] == "gauge" {

			a, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
			if err != nil {
				logrus.Error("Error parsing gauge value: ", err)
			}
			j = append(j, JSONMetrics{
				ID:    k,
				MType: strings.Split(reflect.ValueOf(v).Type().String(), ".")[1],
				Value: &a,
			})
		}
		if strings.Split(reflect.ValueOf(v).Type().String(), ".")[1] == "counter" {
			i, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
			if err != nil {
				logrus.Error("Error parsing counter value: ", err)
			}
			j = append(j, JSONMetrics{
				ID:    k,
				MType: strings.Split(reflect.ValueOf(v).Type().String(), ".")[1],
				Delta: &i,
			})
		}
	}
	return j, nil
}

func (m *MetricsStore) ChangeMetrics(memStats runtime.MemStats) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.PollCount++
	m.MM["BuckHashSys"] = gauge(memStats.BuckHashSys)
	m.MM["Frees"] = gauge(memStats.Frees)
	m.MM["GCCPUFraction"] = gauge(memStats.GCCPUFraction)
	m.MM["GCSys"] = gauge(memStats.GCSys)
	m.MM["HeapAlloc"] = gauge(memStats.HeapAlloc)
	m.MM["Alloc"] = gauge((memStats.Alloc))
	m.MM["HeapIdle"] = gauge(memStats.HeapIdle)
	m.MM["HeapInuse"] = gauge(memStats.HeapInuse)
	m.MM["HeapObjects"] = gauge(memStats.HeapObjects)
	m.MM["HeapReleased"] = gauge(memStats.HeapReleased)
	m.MM["HeapSys"] = gauge(memStats.HeapSys)
	m.MM["LastGC"] = gauge(memStats.LastGC)
	m.MM["Lookups"] = gauge(memStats.Lookups)
	m.MM["MCacheInuse"] = gauge(memStats.MCacheInuse)
	m.MM["MCacheSys"] = gauge(memStats.MCacheSys)
	m.MM["MSpanInuse"] = gauge(memStats.MSpanInuse)
	m.MM["MSpanSys"] = gauge(memStats.MSpanSys)
	m.MM["Mallocs"] = gauge(memStats.Mallocs)
	m.MM["NextGC"] = gauge(memStats.NextGC)
	m.MM["NumForcedGC"] = gauge(memStats.NumForcedGC)
	m.MM["NumGC"] = gauge(memStats.NumGC)
	m.MM["OtherSys"] = gauge(memStats.OtherSys)
	m.MM["PauseTotalNs"] = gauge(memStats.PauseTotalNs)
	m.MM["StackInuse"] = gauge(memStats.StackInuse)
	m.MM["StackSys"] = gauge(memStats.StackSys)
	m.MM["Sys"] = gauge(memStats.Sys)
	m.MM["TotalAlloc"] = gauge(memStats.TotalAlloc)
	m.MM["RandomValue"] = gauge(rand.Float64())
	m.MM["PollCount"] = counter(m.PollCount)
	return nil
}

func (m *MetricsStore) ChangeMetric(nameMet string, value interface{}) error {
	// if strings.Split(reflect.ValueOf(value).Type().String(), ".")[1] == "gauge" {
	m.mux.Lock()
	defer m.mux.Unlock()
	// if err == nil {
	// if _, ok := metrics[nameMet]; ok {
	m.MM[nameMet] = value
	// } else {

	// }
	return nil
}

func (m *MetricsStore) GetMetrics() map[string]interface{} {
	m.mux.Lock()
	defer m.mux.Unlock()
	values := make(map[string]interface{}, (len(m.MM)))
	for k, v := range m.MM {
		values[k] = v
	}
	return values
}
