package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	"github.com/fatih/structs"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

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
	termEnvFlags()
	fmt.Println(config.ArgsM.Key)
	ctx, cancel := context.WithCancel(context.Background())
	go waitSignals(cancel)
	go UpdateMetrics(ctx, config.ArgsM.PollInterval)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Agent is down send metrics.")
			return
		case <-time.After(config.ArgsM.PollInterval):
			// logrus.Info(config.ArgsM.Key)
			err := sendMetricsJSON(ctx, config.ArgsM.Address, []byte(config.ArgsM.Key))
			if err != nil {
				logrus.Error("Error sending POST: ", err)
			}
		}
	}

}

func termEnvFlags() {
	flag.StringVar(&config.FlagsAgent.Address, "a", "127.0.0.1:8080", "Address")
	flag.StringVar(&config.FlagsAgent.Key, "k", "", "Secret key")
	flag.DurationVar(&config.FlagsAgent.ReportInterval, "r", 10000000000, "Report interval")
	flag.DurationVar(&config.FlagsAgent.PollInterval, "p", 2000000000, "Poll interval")

	flag.Parse()

	env := config.LoadConfig()
	envADDR, _ := os.LookupEnv("ADDRESS")
	if envADDR == "" {
		config.ArgsM.Address = config.FlagsAgent.Address
	} else {
		config.ArgsM.Address = env.Address
	}
	envRest, _ := os.LookupEnv("REPORT_INTERVAL")
	if envRest == "" {
		config.ArgsM.ReportInterval = config.FlagsAgent.ReportInterval
	} else {
		config.ArgsM.ReportInterval = env.ReportInterval
	}
	envStoreint, _ := os.LookupEnv("POLL_INTERVAL")
	if envStoreint == "" {
		config.ArgsM.PollInterval = config.FlagsAgent.PollInterval
	} else {
		config.ArgsM.PollInterval = env.PollInterval
	}
	envKey, _ := os.LookupEnv("KEY")
	if envKey == "" {
		config.ArgsM.Key = config.FlagsAgent.Key
	} else {
		config.ArgsM.Key = env.Key
	}

	// fmt.Println(config.ArgsM.Key)
}

func sendMetricsJSON(ctx context.Context, address string, key []byte) error {
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
			met := JSONMetrics[i]
			if string(key) != "" {
				err := SaveHash(&met, []byte(key))
				if err != nil {
					logrus.Error("Error save hash of metrics: ", err)
				}

			}
			logrus.Info("[[[[[[[[[[[[[[[[[[[", met)
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			err = encoder.Encode(met)
			if err != nil {
				logrus.Error(err)
			}
			// var b bytes.Buffer
			// gz, _ := gzip.NewWriterLevel(&b, gzip.BestSpeed)

			// gz.Write(buf.Bytes())
			// gz.Close()
			_, err = client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(&buf).
				Post("http://" + address + "/update/")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SaveHash(JSONMetric *storage.JSONMetrics, key []byte) error {

	switch JSONMetric.MType {
	case "counter":
		data := (fmt.Sprintf("%s:counter:%d", JSONMetric.ID, *JSONMetric.Delta))
		logrus.Info(data)
		h := hmac.New(sha256.New, key)
		h.Write([]byte(data))
		JSONMetric.Hash = string(h.Sum(nil))
	case "gauge":
		data := (fmt.Sprintf("%s:gauge:%f", JSONMetric.ID, *JSONMetric.Value))
		logrus.Info(data)
		h := hmac.New(sha256.New, key)
		h.Write([]byte(data))
		// hh := h.Sum(nil)
		JSONMetric.Hash = fmt.Sprintf("%x", h.Sum(nil))
	}

	return nil
}

//sending metrics to server
// func sendMetrics(ctx context.Context, address string) error {
// 	client := resty.New()
// 	// client.
// 	// SetRetryCount(1).
// 	// SetRetryWaitTime(1 * time.Second).
// 	// SetRetryMaxWaitTime(2 * time.Second)

// 	metrics := storageM.GetMetrics()
// 	for k, v := range metrics {
// 		select {
// 		case <-ctx.Done():
// 			logrus.Info("Send metrics in map ending!")
// 			return nil
// 		default:
// 			typeMet := strings.Split(reflect.ValueOf(v).Type().String(), ".")[1]
// 			value := fmt.Sprintf("%v", v)
// 			_, err := client.R().SetPathParams(map[string]string{
// 				"type": typeMet, "value": value, "name": k,
// 			}).Post("http://" + address + "/update/{type}/{name}/{value}")
// 			if err != nil {
// 				logrus.Error(err)
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

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
			// sendMetrics(ctx)
			storageM.ChangeMetrics(memStats)
		}
	}
}
