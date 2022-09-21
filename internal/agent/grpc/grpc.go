package grpc

import (
	"context"
	"fmt"
	"log"

	pb "github.com/AlekseyKas/metrics/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func cli() {
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	c := pb.NewMetricsClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response, _ := c.GetallMetricsJSON(ctx, &pb.GetMetricsJSONRequest{})
	fmt.Println(response.Result)
}
