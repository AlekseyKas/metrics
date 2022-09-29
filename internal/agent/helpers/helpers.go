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
	"go.uber.org/zap"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/crypto"
	"github.com/AlekseyKas/metrics/internal/storage"
)

// Send metrics to server
func SendMetrics(ctx context.Context, wg *sync.WaitGroup, logger *zap.Logger, pubKey string, storageM storage.StorageAgent) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Agent is down send metrics.")
			return
		case <-time.After(config.ArgsM.PollInterval):
			err := SendMetricsSlice(ctx, logger, config.ArgsM.Address, pubKey, []byte(config.ArgsM.Key), storageM)
			if err != nil {
				logger.Error("Error sending POST: ", zap.Error(err))
			}
		}
	}
}

// Prepare and sending metrics to server
func SendMetricsSlice(ctx context.Context, logger *zap.Logger, address string, pubKey string, key []byte, storageM storage.StorageAgent) error {
	client := resty.New()

	JSONMetrics, err := storageM.GetMetricsJSON()
	if err != nil {
		logger.Error("Error getting metrics json format", zap.Error(err))
	}

	for i := 0; i < len(JSONMetrics); i++ {
		select {
		case <-ctx.Done():
			logger.Info("Send metrics in map ending!")
			return nil
		default:
			if string(key) != "" {
				_, err = SaveHash(&JSONMetrics[i], []byte(key))
				if err != nil {
					logger.Error("Error save hash of metrics: ", zap.Error(err))
				}
			}
		}
	}
	var buf bytes.Buffer
	var b bytes.Buffer

	encoder := json.NewEncoder(&buf)
	err = encoder.Encode(&JSONMetrics)
	if err != nil {
		logger.Error("Error encoding JSON metrics", zap.Error(err))
	}

	gz, _ := gzip.NewWriterLevel(&b, gzip.BestSpeed)

	_, err = gz.Write(buf.Bytes())
	if err != nil {
		logger.Error("Error write gz metrics: ", zap.Error(err))
	}
	gz.Close()
	// Encryption
	if pubKey != "" {

		var data []byte
		data, err = crypto.EncryptData(b.Bytes(), pubKey)
		if err != nil {
			logger.Error("Error encrypt data: ", zap.Error(err))
		}
		_, err = client.R().
			SetHeader("X-Real-IP", "127.0.0.1").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetBody(data).
			Post("http://" + address + "/updates/")
		if err != nil {
			return err
		}
		return nil
	} else {
		_, err = client.R().
			SetHeader("X-Real-IP", "127.0.0.1").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetBody(&b).
			Post("http://" + address + "/updates/")
		if err != nil {
			return err
		}
		return nil
	}
}

// Set sha256 hash for metric
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

//Update metrics terminating
func UpdateMetrics(ctx context.Context, pollInterval time.Duration, wg *sync.WaitGroup, logger *zap.Logger, storageM storage.StorageAgent) {
	defer wg.Done()
	for {
		select {
		//send command to ending
		case <-ctx.Done():
			logger.Info("Agent is down update metrics!")
			return
		case <-time.After(pollInterval):
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			err := storageM.ChangeMetrics(memStats)
			if err != nil {
				logger.Error("Error changing metrics ChangeMetrics: ", zap.Error(err))
			}
		}
	}
}

//Update new metrics
func UpdateMetricsNew(ctx context.Context, pollInterval time.Duration, wg *sync.WaitGroup, logger *zap.Logger, storageM storage.StorageAgent) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Agent is down update metrics mem & cpu!")
			return
		case <-time.After(pollInterval):
			mem, err := mem.VirtualMemory()
			if err != nil {
				logger.Error("Error get metic Virtualmemory: ", zap.Error(err))
			}
			cpu, err := cpu.Percent(time.Second, false)
			if err != nil {
				logger.Error("Error get metric cpu cover: ", zap.Error(err))
			}
			err = storageM.ChangeMetricsNew(mem, cpu)
			if err != nil {
				logger.Error("Error change new metrics ChangeMetricsNew: ", zap.Error(err))
			}
		}
	}
}

// Wait siglans SIGTERM, SIGINT, SIGQUIT
func WaitSignals(cancel context.CancelFunc, logger *zap.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-terminate
	logger.Info("Agent is down terminate!")
	cancel()
}
