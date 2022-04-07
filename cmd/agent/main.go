package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const pollInterval = 2 * time.Second
const reportInterval = 10 * time.Second

var storageM storage.StorageAgent

func SetStorageAgent(s storage.StorageAgent) {
	storageM = s
}

func main() {
	var MapMetrics map[string]interface{} = structs.Map(storage.Metrics{})
	//инициализация хранилища метрик
	s := &storage.MetricsStore{
		MM: MapMetrics,
	}
	SetStorageAgent(s)

	ctx, cancel := context.WithCancel(context.Background())
	go waitSignals(cancel)
	go UpdateMetrics(ctx, pollInterval)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Agent is down send metrics.")
			return
		case <-time.After(reportInterval):
			err := sendMetricsJSON(ctx)
			if err != nil {
				logrus.Error("Error sending POST: ", err)
			}
		}
	}

}
func sendMetricsJSON(ctx context.Context) error {
	client := resty.New()
	client.
		SetRetryCount(1).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(2 * time.Second)

	jsonMetrics, err := storageM.GetMetricsJSON()
	if err != nil {
		logrus.Error("Error getting metrics json format", err)
	}

	for i := 0; i < len(jsonMetrics); i++ {
		select {
		case <-ctx.Done():
			logrus.Info("Send metrics in map ending!")
			return nil
		default:
			out, err := json.Marshal(jsonMetrics[i])
			if err != nil {
				logrus.Error("Error marshaling metric: ", err)
			}

			_, err = client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(out).
				Post("http://127.0.0.1:8080/update/")
			if err != nil {
				return err
			}
			// fmt.Println(string(out))
		}
	}
	return nil
}

//sending metrics to server
func sendMetrics(ctx context.Context) error {
	client := resty.New()
	client.
		SetRetryCount(1).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(2 * time.Second)

	metrics := storageM.GetMetrics()
	for k, v := range metrics {
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
		case <-time.After(pollInterval):
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			storageM.ChangeMetrics(memStats)
		}
	}
}
