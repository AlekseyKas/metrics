package grpc

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	pb "github.com/AlekseyKas/metrics/proto"
	"github.com/fatih/structs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPC(t *testing.T) {
	s := &storage.MetricsStore{
		MM:  structs.Map(storage.Metrics{}),
		Ctx: context.Background(),
	}
	logger := GRPCSrv.Logger
	config.TermEnvFlags()
	srv := New(s.Ctx, logger, s, config.ArgsM)
	go func() {
		err := srv.Start()
		log.Fatal(err)
		time.Sleep(10 * time.Second)

	}()
	//client for test
	c := runClient(config.ArgsM.Address)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := c.GetMetricsJSON(ctx, &pb.Empty{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("==================================%+v\n", resp)
	metric := &pb.Metric{
		ID:    "Alloc",
		MType: "gauge",
	}
	data, _ := c.GetMetricData(ctx, metric)
	fmt.Println("==================================1", data)

	m := &pb.MetricData{
		ID:    "Alloc",
		Mtype: "gauge",
		Data:  "0.22",
	}
	_, err = c.UpdateMetric(ctx, m)
	if err != nil {
		log.Fatal(err)
	}

	dat, _ := c.GetMetricData(ctx, metric)
	fmt.Println("###################################", dat)
}

func runClient(address string) pb.MetricsMClient {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	c := pb.NewMetricsMClient(conn)
	return c
}
