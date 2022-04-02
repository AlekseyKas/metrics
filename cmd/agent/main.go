package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

//init typs
type gauge float64
type counter int64

//delay
const pollInterval = 2 * time.Second
const reportInterval = 10 * time.Second

//Struct for metrics
type Metrics struct {
	mux       sync.Mutex
	MM        map[string]interface{}
	PollCount int
}

var M Metrics = Metrics{MM: make(map[string]interface{})}

func (ptr *Metrics) Get() (values map[string]interface{}) {
	ptr.mux.Lock()
	defer ptr.mux.Unlock()
	values = make(map[string]interface{})
	for k, v := range ptr.MM {
		values[k] = v
	}
	return
}

func (ptr *Metrics) Update() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	ptr.mux.Lock()
	defer ptr.mux.Unlock()
	ptr.PollCount++
	ptr.MM["Alloc"] = gauge((memStats.Alloc))
	ptr.MM["BuckHashSys"] = gauge(memStats.BuckHashSys)
	ptr.MM["Frees"] = gauge(memStats.Frees)
	ptr.MM["GCCPUFraction"] = gauge(memStats.GCCPUFraction)
	ptr.MM["GCSys"] = gauge(memStats.GCSys)
	ptr.MM["HeapAlloc"] = gauge(memStats.HeapAlloc)
	ptr.MM["HeapIdle"] = gauge(memStats.HeapIdle)
	ptr.MM["HeapInuse"] = gauge(memStats.HeapInuse)
	ptr.MM["HeapObjects"] = gauge(memStats.HeapObjects)
	ptr.MM["HeapReleased"] = gauge(memStats.HeapReleased)
	ptr.MM["HeapSys"] = gauge(memStats.HeapSys)
	ptr.MM["LastGC"] = gauge(memStats.LastGC)
	ptr.MM["Lookups"] = gauge(memStats.Lookups)
	ptr.MM["MCacheInuse"] = gauge(memStats.MCacheInuse)
	ptr.MM["MCacheSys"] = gauge(memStats.MCacheSys)
	ptr.MM["MSpanInuse"] = gauge(memStats.MSpanInuse)
	ptr.MM["MSpanSys"] = gauge(memStats.MSpanSys)
	ptr.MM["Mallocs"] = gauge(memStats.Mallocs)
	ptr.MM["NextGC"] = gauge(memStats.NextGC)
	ptr.MM["NumForcedGC"] = gauge(memStats.NumForcedGC)
	ptr.MM["NumGC"] = gauge(memStats.NumGC)
	ptr.MM["OtherSys"] = gauge(memStats.OtherSys)
	ptr.MM["PauseTotalNs"] = gauge(memStats.PauseTotalNs)
	ptr.MM["StackInuse"] = gauge(memStats.StackInuse)
	ptr.MM["StackSys"] = gauge(memStats.StackSys)
	ptr.MM["Sys"] = gauge(memStats.Sys)
	ptr.MM["TotalAlloc"] = gauge(memStats.TotalAlloc)
	ptr.MM["RandomValue"] = gauge(rand.Float64())
	ptr.MM["PollCount"] = counter(ptr.PollCount)
}

func main() {
	M.Update()

	ctx, cancel := context.WithCancel(context.Background())
	go waitSignals(cancel)
	go UpdateMetrics(ctx, pollInterval)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Agent is down send metrics.")
			return
		case <-time.After(reportInterval):
			err := saveMetrics(ctx)
			if err != nil {
				logrus.Error("Error sending POST: ", err)
			}
		}
	}

}

//sending metrics to server
func saveMetrics(ctx context.Context) error {
	client := resty.New()
	client.
		SetRetryCount(1).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(2 * time.Second)

	for k, v := range M.Get() {
		select {
		case <-ctx.Done():
			logrus.Info("Send metrics in map ending!")
			return nil
		default:
			typeMet := strings.Split(reflect.ValueOf(v).Type().String(), ".")[1]
			value := fmt.Sprintf("%v", v)
			_, err := client.R().SetPathParams(map[string]string{
				"type": typeMet, "value": value, "name": k,
			}).Post("http://127.0.0.1:8080/update/{type}/{name}/{value}")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//wating signals
func waitSignals(cancel context.CancelFunc) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		sig := <-terminate
		switch sig {
		case os.Interrupt:
			logrus.Info("Agent is down terminate!")
			cancel()
			return
		}
	}
}

//Update metrics
func UpdateMetrics(ctx context.Context, pollInterval time.Duration) {
	for {
		select {
		//send command to ending
		case <-ctx.Done():
			logrus.Info("Agent is down update metrics!")
			return
		case <-time.After(time.Second * pollInterval):
			M.Update()
		}
	}
}
