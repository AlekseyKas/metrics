package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/structs"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
)

var storageM storage.StorageAgent
var wg sync.WaitGroup

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
	// fmt.Println(config.ArgsM.Key)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go waitSignals(cancel)
	wg.Add(1)
	go UpdateMetrics(ctx, config.ArgsM.PollInterval)
	wg.Add(1)
	go SendMetrics(ctx)
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		logrus.Info("Agent is down send metrics.")
	// 		return
	// 	case <-time.After(config.ArgsM.PollInterval):
	// 		err := sendMetricsSlice(ctx, config.ArgsM.Address, []byte(config.ArgsM.Key))
	// 		if err != nil {
	// 			logrus.Error("Error sending POST: ", err)
	// 		}
	// 	}
	// }
	wg.Wait()

}

func SendMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Agent is down send metrics.")
			wg.Done()
			return
		case <-time.After(config.ArgsM.PollInterval):
			err := sendMetricsSlice(ctx, config.ArgsM.Address, []byte(config.ArgsM.Key))
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
}
func sendMetricsSlice(ctx context.Context, address string, key []byte) error {
	client := resty.New()

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
			if string(key) != "" {
				_, err := SaveHash(&JSONMetrics[i], []byte(key))
				if err != nil {
					logrus.Error("Error save hash of metrics: ", err)
				}
			}
			// if JSONMetrics[i].ID == "PollCount" {
			// 	logrus.Info("oooooooo", *JSONMetrics[i].Delta, ": ", JSONMetrics[i].Hash)

			// }
		}
	}

	var buf bytes.Buffer
	var b bytes.Buffer

	encoder := json.NewEncoder(&buf)
	err = encoder.Encode(&JSONMetrics)
	if err != nil {
		logrus.Error(err)
	}

	gz, _ := gzip.NewWriterLevel(&b, gzip.BestSpeed)

	gz.Write(buf.Bytes())
	gz.Close()
	_, err = client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(&b).
		Post("http://" + address + "/updates/")
	if err != nil {
		return err
	}
	return nil
}

func SaveHash(JSONMetric *storage.JSONMetrics, key []byte) (hash string, err error) {
	var hh string
	switch JSONMetric.MType {
	case "counter":
		data := (fmt.Sprintf("%s:counter:%d", JSONMetric.ID, *JSONMetric.Delta))
		h := hmac.New(sha256.New, key)
		h.Write([]byte(data))
		JSONMetric.Hash = fmt.Sprintf("%x", h.Sum(nil))
		hh = fmt.Sprintf("%x", h.Sum(nil))
	case "gauge":
		data := (fmt.Sprintf("%s:gauge:%f", JSONMetric.ID, *JSONMetric.Value))
		h := hmac.New(sha256.New, key)
		h.Write([]byte(data))
		JSONMetric.Hash = fmt.Sprintf("%x", h.Sum(nil))
		hh = fmt.Sprintf("%x", h.Sum(nil))
	}
	return hh, nil
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
			wg.Done()
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
			wg.Done()
			return
		case <-time.After(pollInterval):
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			storageM.ChangeMetrics(memStats)
		}
	}
}
