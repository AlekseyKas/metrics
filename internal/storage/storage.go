package storage

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/AlekseyKas/metrics/internal/config"
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

// type ValidStruct struct {
// 	ID    interface{} `json:"id"`              // имя метрики
// 	MType interface{} `json:"type"`            // параметр, принимающий значение gauge или counter
// 	Delta interface{} `json:"delta,omitempty"` // значение метрики в случае передачи counter
// 	Value interface{} `json:"value,omitempty"` // значение метрики в случае передачи gauge
// }

type StorageAgent interface {
	GetMetrics() map[string]interface{}
	ChangeMetrics(metrics runtime.MemStats) error

	GetMetricsJSON() ([]JSONMetrics, error)
}

type Storage interface {
	GetMetrics() map[string]interface{}
	ChangeMetric(nameMet string, value interface{}, params config.Param) error
	GetStructJSON() JSONMetrics
	LoadMetricsFile(file []byte)
	GetMetricsJSON() ([]JSONMetrics, error)
}

// func (m *MetricsStore) GetMetricsJSON() []JSONMetrics {
// 	var jMetric JSONMetrics
// 	var metrics []JSONMetrics
// 	for k, v := range m.MM {
// 		switch strings.Split(reflect.ValueOf(v).Type().String(), ".")[1] {
// 		case "counter":
// 			i, err := strconv.Atoi(fmt.Sprintf("%v", v))
// 			if err != nil {
// 				logrus.Error("Error convert int counter")
// 			}
// 			ii := int64(i)
// 			jMetric = JSONMetrics{
// 				ID:    k,
// 				MType: "counter",
// 				Delta: &ii,
// 			}
// 			metrics = append(metrics, jMetric)
// 		case "gauge":
// 			f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
// 			if err != nil {
// 				logrus.Error("Error convert int float")
// 			}
// 			jMetric = JSONMetrics{
// 				ID:    k,
// 				MType: "counter",
// 				Value: &f,
// 			}
// 			metrics = append(metrics, jMetric)
// 		}
// 	}
// 	return metrics
// }

func (m *MetricsStore) LoadMetricsFile(file []byte) {
	m.mux.Lock()
	defer m.mux.Unlock()
	var jMetric []JSONMetrics

	err := json.Unmarshal(file, &jMetric)
	if err != nil {
		logrus.Error("Error unmarshaling file to map", err)
	}
	for i := 0; i < len(jMetric); i++ {
		if _, ok := m.MM[jMetric[i].ID]; ok {
			for k, _ := range m.MM {
				if k == jMetric[i].ID {
					if jMetric[i].Delta != nil {
						v := jMetric[i].Delta
						m.MM[k] = counter(*v)
					}
					if jMetric[i].Value != nil {
						v := jMetric[i].Value
						m.MM[k] = gauge(*v)
					}
				}
			}
		} else {
			ty := jMetric[i].MType
			switch ty {
			case "counter":
				m.MM[jMetric[i].ID] = counter(*jMetric[i].Delta)
			case "gauge":
				m.MM[jMetric[i].ID] = gauge(*jMetric[i].Delta)
			}
		}
	}
}

func (m *MetricsStore) GetStructJSON() JSONMetrics {
	s := JSONMetrics{}
	return s
}

func (m *MetricsStore) GetMetricsJSON() ([]JSONMetrics, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
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

func (m *MetricsStore) ChangeMetric(nameMet string, value interface{}, params config.Param) error {
	sl, err := m.GetMetricsJSON()
	if err != nil {
		logrus.Error(err)
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	if params.StoreInterval == 0 {
		m.MM[nameMet] = value
		file, err := os.OpenFile(params.StoreFile, os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			logrus.Error("Error open file for writing:!!!!!!!! ", err)
		}
		defer file.Close()

		data, err := json.Marshal(sl)
		if err != nil {
			logrus.Error("Error marshaling metrics : ", err)
		}
		file.Write(data)
	} else {
		// m.mux.Lock()
		// defer m.mux.Unlock()
		m.MM[nameMet] = value
	}

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
