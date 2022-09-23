package grpc

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AlekseyKas/metrics/internal/agent/helpers"
	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	pb "github.com/AlekseyKas/metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// func client() {
// 	conn, err := grpc.Dial("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	c := pb.NewMetricsClient(conn)
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	response, _ := c.GetallMetricsJSON(ctx, &pb.GetMetricsJSONRequest{})
// 	fmt.Println(string(response.Result))
// }

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

	JSONMetrics, err := storageM.GetMetricsJSON()
	fmt.Println(JSONMetrics[0])
	if err != nil {
		logger.Error("Error getting metrics json format", zap.Error(err))
	}
	grpcMetrics := make([]*pb.JSONMetrics, 0)
	// grpc client
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	c := pb.NewMetricsMClient(conn)

	for _, v := range JSONMetrics {
		select {
		case <-ctx.Done():
			logger.Info("Send metrics in map ending!")
			return nil
		default:
			if string(key) != "" {
				_, err = helpers.SaveHash(&v, []byte(key))
				if err != nil {
					logger.Error("Error save hash of metrics: ", zap.Error(err))
				}
			}

			grpcMetrics = append(grpcMetrics, converToGRPC(v))
		}
	}
	// fmt.Println("====================", grpcMetrics)
	client()
	_, err = c.SendMetricsJSON(ctx, &pb.SendMetricsJSONRequest{JSONMetrics: grpcMetrics})
	if err != nil {
		logger.Error("Error sending metrics: ", zap.Error(err))
	}
	return nil
}

func client() {
	conn, err := grpc.Dial("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	c := pb.NewMetricsMClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// response, _ := c.GetMetricsJSON(ctx, &pb.Empty{})
	// fmt.Println("==================================", response)
	// metric := &pb.Metric{
	// 	ID:    "Alloc",
	// 	MType: "gauge",
	// }
	// data, _ := c.GetMetricData(ctx, metric)
	// fmt.Println("==================================", data)
	m := &pb.MetricData{
		ID:    "Alloc",
		Mtype: "gauge",
		Data:  "0.22",
	}
	_, err = c.UpdateMetric(ctx, m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("==================================")

}

func converToGRPC(s storage.JSONMetrics) (result *pb.JSONMetrics) {

	result = &pb.JSONMetrics{
		ID:    s.ID,
		MType: s.MType,
		Hash:  s.Hash,
	}
	if s.Delta != nil {
		result.Delta = *s.Delta
	}
	if s.Value != nil {
		result.Value = *s.Value
	}

	return result
}
