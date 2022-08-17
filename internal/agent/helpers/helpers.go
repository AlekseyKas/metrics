package helpers

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
)

// Send metrics to server
func SendMetrics(ctx context.Context, wg *sync.WaitGroup, storageM storage.StorageAgent) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Agent is down send metrics.")
			return
		case <-time.After(config.ArgsM.PollInterval):
			err := SendMetricsSlice(ctx, config.ArgsM.Address, []byte(config.ArgsM.Key), storageM)
			if err != nil {
				logrus.Error("Error sending POST: ", err)
			}
		}
	}
}

// Prepare and sending metrics to server
func SendMetricsSlice(ctx context.Context, address string, key []byte, storageM storage.StorageAgent) error {
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
				_, err := saveHash(&JSONMetrics[i], []byte(key))
				if err != nil {
					logrus.Error("Error save hash of metrics: ", err)
				}
			}
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

	_, err = gz.Write(buf.Bytes())
	if err != nil {
		logrus.Error("Error write gz metrics: ", err)
	}
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

// Set sha256 hash for metric
func saveHash(JSONMetric *storage.JSONMetrics, key []byte) (hash string, err error) {
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

//Update metrics terminating
func UpdateMetrics(ctx context.Context, pollInterval time.Duration, wg *sync.WaitGroup, storageM storage.StorageAgent) {
	defer wg.Done()
	for {
		select {
		//send command to ending
		case <-ctx.Done():
			logrus.Info("Agent is down update metrics!")
			return
		case <-time.After(pollInterval):
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			err := storageM.ChangeMetrics(memStats)
			if err != nil {
				logrus.Error("Error changing metrics ChangeMetrics: ", err)
			}
		}
	}
}

//Update new metrics
func UpdateMetricsNew(ctx context.Context, pollInterval time.Duration, wg *sync.WaitGroup, storageM storage.StorageAgent) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Agent is down update metrics mem & cpu!")
			return
		case <-time.After(pollInterval):
			mem, err := mem.VirtualMemory()
			if err != nil {
				logrus.Error(err)
			}
			cpu, err := cpu.Percent(time.Second, false)
			if err != nil {
				logrus.Error(err)
			}
			err = storageM.ChangeMetricsNew(mem, cpu)
			if err != nil {
				logrus.Error("Error change new metrics ChangeMetricsNew: ", err)
			}
		}
	}
}

// Wait siglans SIGTERM, SIGINT, SIGQUIT
func WaitSignals(cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-terminate
	logrus.Info("Agent is down terminate!")
	cancel()
}
