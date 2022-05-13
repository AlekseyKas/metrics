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

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"

	"github.com/AlekseyKas/metrics/cmd/server/database"
	"github.com/AlekseyKas/metrics/internal/config"
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
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type MetricsStore struct {
	mux       sync.Mutex
	MM        map[string]interface{}
	PollCount int
}

type StorageAgent interface {
	GetMetrics() map[string]interface{}
	ChangeMetrics(metrics runtime.MemStats) error
	ChangeMetricsNew(metrics *mem.VirtualMemoryStat, cpu []float64) error
	GetMetricsJSON() ([]JSONMetrics, error)
}

type Storage interface {
	InitDB(jm []JSONMetrics) error
	LoadMetricsDB() error
	ChangeMetricDB(nameMet string, value interface{}, typeMet string, params config.Args) error
	GetMetrics() map[string]interface{}
	ChangeMetric(nameMet string, value interface{}, params config.Args) error
	GetStructJSON() JSONMetrics
	LoadMetricsFile(file []byte)
	GetMetricsJSON() ([]JSONMetrics, error)
	GetSliceStruct() []JSONMetrics
}

func (m *MetricsStore) LoadMetricsDB() error {
	var id string
	var metricType string
	var value *float64
	var delta *int64
	row, err := database.Conn.Query("SELECT id, metric_type, value, delta FROM metrics")
	if err != nil {
		logrus.Error("Error select all from table metrics: ", err)
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	for row.Next() {
		err = row.Scan(&id, &metricType, &value, &delta)
		if err != nil {
			logrus.Error("Error scan row in select all: ", err)
		}
		if metricType == "gauge" {
			m.MM[id] = value
		} else {
			m.MM[id] = delta
		}
	}
	return nil
}

//update metric in database
func (m *MetricsStore) ChangeMetricDB(nameMet string, value interface{}, typeMet string, params config.Args) error {
	if params.DBURL != "" {
		switch typeMet {
		case "gauge":
			_, err := database.Conn.Exec("INSERT INTO metrics (id, metric_type, value) VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET value = $3", nameMet, typeMet, value)
			if err != nil {
				logrus.Error("Error insert metric gauge in database: ", err)
			}
		case "counter":
			_, err := database.Conn.Exec("INSERT INTO metrics (id, metric_type, delta) VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET delta = $3", nameMet, typeMet, value)
			if err != nil {
				logrus.Error("Error insert metric counter in database: ", err)
			}
		}
	}
	return nil
}

//init database if don't exist table
func (m *MetricsStore) InitDB(jm []JSONMetrics) error {

	_, err := database.Conn.Exec("CREATE TABLE IF NOT EXISTS metrics (id VARCHAR NOT NULL UNIQUE, metric_type VARCHAR NOT NULL, delta BIGINT, value DOUBLE PRECISION)")
	if err != nil {
		logrus.Error("Error create table: ", err)
	}
	return nil
}

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
			for k := range m.MM {
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
				m.MM[jMetric[i].ID] = gauge(*jMetric[i].Value)
			}
		}
	}
}

func (m *MetricsStore) GetStructJSON() JSONMetrics {
	s := JSONMetrics{}
	return s
}

func (m *MetricsStore) GetSliceStruct() []JSONMetrics {
	s := []JSONMetrics{}
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

func (m *MetricsStore) ChangeMetricsNew(mem *mem.VirtualMemoryStat, cpu []float64) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	logrus.Info("mem: ", mem, "cpu: ", cpu)
	m.MM["TotalMemory"] = gauge(float64(mem.Total))
	m.MM["FreeMemory"] = gauge(float64(mem.Free))
	m.MM["CPUutilization1"] = gauge(cpu[0])
	return nil
}

func (m *MetricsStore) ChangeMetric(nameMet string, value interface{}, params config.Args) error {
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
