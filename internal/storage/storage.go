package storage

import (
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sync"
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

type MetricsStore struct {
	mux       sync.Mutex
	MM        map[string]interface{}
	PollCount int
}

type StorageAgent interface {
	GetMetrics() map[string]interface{}
	ChangeMetrics(metrics runtime.MemStats) error
}

type Storage interface {
	GetMetrics() map[string]interface{}
	ChangeGauge(nameMet string, value interface{}) error
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

func (m *MetricsStore) ChangeGauge(nameMet string, value interface{}) error {
	if strings.Split(reflect.ValueOf(value).Type().String(), ".")[1] == "gauge" {
		m.mux.Lock()
		defer m.mux.Unlock()
		m.MM[nameMet] = value
	} else {
		m.mux.Lock()
		defer m.mux.Unlock()
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
