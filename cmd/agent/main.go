package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/fatih/structs"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// const pollInterval = 2 * time.Second
// const reportInterval = 10 * time.Second

type Param struct {
	POLL_INTERVAL   time.Duration `env:"POLL_INTERVAL"`
	REPORT_INTERVAL time.Duration `env:"REPORT_INTERVAL"`
	ADDRESS         string        `env:"ADDRESS"`
}

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
	p := GetParam()
	ctx, cancel := context.WithCancel(context.Background())
	go waitSignals(cancel)
	go UpdateMetrics(ctx, p.POLL_INTERVAL)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Agent is down send metrics.")
			return
		case <-time.After(p.REPORT_INTERVAL):
			err := sendMetricsJSON(ctx, p.ADDRESS)
			if err != nil {
				logrus.Error("Error sending POST: ", err)
			}
		}
	}

}

//get param from env
func GetParam() Param {
	var param Param
	err := env.Parse(&param)
	if err != nil {
		log.Fatal(err)
	}
	if param.POLL_INTERVAL == 0 {
		param.POLL_INTERVAL = 2 * time.Second
	}
	if param.REPORT_INTERVAL == 0 {
		param.REPORT_INTERVAL = 10 * time.Second
	}
	if param.ADDRESS == "" {
		param.ADDRESS = "127.0.0.1:8080"
	}
	return param
}
func sendMetricsJSON(ctx context.Context, address string) error {
	client := resty.New()
	// client.
	// 	SetRetryCount(1).
	// 	SetRetryWaitTime(1 * time.Second).
	// 	SetRetryMaxWaitTime(2 * time.Second)

	JSONMetrics, err := storageM.GetMetricsJSON()
	if err != nil {
		logrus.Error("Error getting metrics json format", err)
	}

	for i := 0; i < len(JSONMetrics); i++ {
		select {
		case <-ctx.Done():
			logrus.Info("Send metrics in map ending!")
			return nil
		default:
			// out, err := json.Marshal(JSONMetrics[i])
			// if err != nil {
			// 	logrus.Error("Error marshaling metric: ", err)
			// }
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			err = encoder.Encode(JSONMetrics[i])
			if err != nil {
				logrus.Error(err)
			}
			logrus.Info(buf.String())

			_, err = client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(buf.Bytes()).
				Post("http://" + address + "/update/")
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
	// client.
	// SetRetryCount(1).
	// SetRetryWaitTime(1 * time.Second).
	// SetRetryMaxWaitTime(2 * time.Second)

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
				logrus.Error(err)
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
			sendMetrics(ctx)
			storageM.ChangeMetrics(memStats)
		}
	}
}
