package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/storage"
	pb "github.com/AlekseyKas/metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Init type grpc server
type grpcServer struct {
	pb.UnimplementedMetricsServer
	Logger   *zap.Logger
	StorageM storage.Storage
	Args     config.Args
}

// Init grpc server.
var grpcSrv grpcServer

// Init type metrics gauge and counter
type gauge float64
type counter int64

// Terminate storage server
func setStorage(s storage.Storage) {
	grpcSrv.StorageM = s
}

// Init grpc server
func New(logger *zap.Logger, storage *storage.MetricsStore, args config.Args) *grpcServer {
	grpcSrv := new(grpcServer)
	grpcSrv.Logger = logger
	setStorage(storage)
	grpcSrv.Args = args
	return grpcSrv
}

// Start grpc server.
func (s *grpcServer) Start() error {
	listen, err := net.Listen("tcp", grpcSrv.Args.Address)
	if err != nil {
		grpcSrv.Logger.Error("Error start grpc server", zap.Error(err))
	}

	srv := grpc.NewServer()
	// Register server.
	pb.RegisterMetricsServer(srv, s)

	fmt.Println("Сервер gRPC начал работу")
	// Get request grpc
	if err := srv.Serve(listen); err != nil {
		grpcSrv.Logger.Error("Error getting request grpc: ", zap.Error(err))
	}
	return err
}

// Get all metrics url /
func (s *grpcServer) GetallMetricsJSON(ctx context.Context, rq *pb.GetMetricsJSONRequest) (*pb.GetMetricsJSONResponse, error) {
	response := new(pb.GetMetricsJSONResponse)
	response.Result = "bla"
	return response, nil
}
