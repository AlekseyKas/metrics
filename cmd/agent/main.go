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
	"syscall"
	"time"

	"github.com/fatih/structs"
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

func main() {
	M := Metrics{}
	//init terminate
	ctx, cancel := context.WithCancel(context.Background())
	go waitSignals(cancel)

	//set ticker dor sending
	// tickerSend := time.NewTicker(reportInterval)
	// defer tickerSend.Stop()
	go UpdateMetrics(ctx, &M, pollInterval)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Agent is down send metrics.")
			return
		default:
			time.Sleep(reportInterval)
			err := saveMetrics(ctx, M)
			if err != nil {
				logrus.Error("Error sending POST: ", err)
			}
		}
	}

}

//sending metrics to server
func saveMetrics(ctx context.Context, M Metrics) error {
	metricsMap := structs.Map(M)

	client := resty.New()
	client.
		SetRetryCount(2).
		SetRetryWaitTime(10 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second)

	for k, v := range metricsMap {
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
				fmt.Println(err)
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
func UpdateMetrics(ctx context.Context, M *Metrics, pollInterval time.Duration) {
	var memStats runtime.MemStats
	//Init ticker
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	//init pollcount
	PollCount := 0
	for {
		select {
		//send command to ending
		case <-ctx.Done():
			logrus.Info("Agent is down update metrics!")
			return
		default:
			<-ticker.C
			runtime.ReadMemStats(&memStats)
			//upper PollCount
			PollCount++
			// fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", PollCount)
			//Update metrics
			*M = Metrics{
				Alloc:         gauge((memStats.Alloc)),
				BuckHashSys:   gauge(memStats.BuckHashSys),
				Frees:         gauge(memStats.Frees),
				GCCPUFraction: gauge(memStats.GCCPUFraction),
				GCSys:         gauge(memStats.GCSys),
				HeapAlloc:     gauge(memStats.HeapAlloc),
				HeapIdle:      gauge(memStats.HeapIdle),
				HeapInuse:     gauge(memStats.HeapInuse),
				HeapObjects:   gauge(memStats.HeapObjects),
				HeapReleased:  gauge(memStats.HeapReleased),
				HeapSys:       gauge(memStats.HeapSys),
				LastGC:        gauge(memStats.LastGC),
				Lookups:       gauge(memStats.Lookups),
				MCacheInuse:   gauge(memStats.MCacheInuse),
				MCacheSys:     gauge(memStats.MCacheSys),
				MSpanInuse:    gauge(memStats.MSpanInuse),
				MSpanSys:      gauge(memStats.MSpanSys),
				Mallocs:       gauge(memStats.Mallocs),
				NextGC:        gauge(memStats.NextGC),
				NumForcedGC:   gauge(memStats.NumForcedGC),
				NumGC:         gauge(memStats.NumGC),
				OtherSys:      gauge(memStats.OtherSys),
				PauseTotalNs:  gauge(memStats.PauseTotalNs),
				StackInuse:    gauge(memStats.StackInuse),
				StackSys:      gauge(memStats.StackSys),
				Sys:           gauge(memStats.Sys),
				TotalAlloc:    gauge(memStats.TotalAlloc),
				RandomValue:   gauge(rand.Float64()),
				PollCount:     counter(PollCount),
			}
		}
	}
}
