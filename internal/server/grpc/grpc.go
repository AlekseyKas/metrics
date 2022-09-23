package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"unsafe"

	"github.com/AlekseyKas/metrics/internal/config"
	"github.com/AlekseyKas/metrics/internal/server/handlers"
	"github.com/AlekseyKas/metrics/internal/storage"
	pb "github.com/AlekseyKas/metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Init type grpc server
type grpcServer struct {
	pb.UnimplementedMetricsMServer
	Logger   *zap.Logger
	StorageM storage.Storage
	Args     config.Args
	Ctx      context.Context
}

// Init grpc server.
var GRPCSrv grpcServer

// Init type metrics gauge and counter
type gauge float64
type counter int64

// Terminate storage server
func setStorage(s storage.Storage) {
	GRPCSrv.StorageM = s
}

// Init grpc server
func New(ctx context.Context, logger *zap.Logger, storage *storage.MetricsStore, args config.Args) *grpcServer {
	srv := new(grpcServer)
	GRPCSrv.Logger = logger
	setStorage(storage)
	GRPCSrv.Args = args
	return srv
}

// Start grpc server.
func (s *grpcServer) Start() error {
	listen, err := net.Listen("tcp", GRPCSrv.Args.Address)
	if err != nil {
		GRPCSrv.Logger.Error("Error start grpc server", zap.Error(err))
	}

	srv := grpc.NewServer()
	// Register server.

	pb.RegisterMetricsMServer(srv, s)

	fmt.Println("Сервер gRPC начал работу")
	// Get request grpc
	if err := srv.Serve(listen); err != nil {
		GRPCSrv.Logger.Error("Error getting request grpc: ", zap.Error(err))
	}
	return err
}

// Get and save metrics
func (srv *grpcServer) SendMetricsJSON(ctx context.Context, req *pb.SendMetricsJSONRequest) (*pb.Empty, error) {
	var err error
	s := convertToSJSONMetrics(req.JSONMetrics)
	var typeMet string
	var nameMet string
	for i := 0; i < len(s); i++ {
		typeMet = s[i].MType
		nameMet = s[i].ID
		metrics := GRPCSrv.StorageM.GetMetrics()

		if GRPCSrv.Args.Key != "" {
			var b bool
			b, err = handlers.CompareHash(&s[i], []byte(config.ArgsM.Key))
			if err != nil {
				GRPCSrv.Logger.Error("Error compare hash of metrics: ", zap.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}
			if b {
				//update gauge
				if typeMet == "gauge" {
					if s[i].Value != nil {
						if metrics[nameMet] != gauge(*s[i].Value) {
							err = GRPCSrv.StorageM.ChangeMetric(nameMet, gauge(*s[i].Value), config.ArgsM)
							if err != nil {
								GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
								return nil, status.Error(codes.Internal, err.Error())
							}
							err = GRPCSrv.StorageM.ChangeMetricDB(nameMet, *s[i].Value, typeMet, config.ArgsM)
							if err != nil {
								GRPCSrv.Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
								return nil, status.Error(codes.Internal, err.Error())
							}
						}
					}
				}
				//update counter
				if typeMet == "counter" {
					var valueMetInt int
					if s[i].Delta != nil {
						if _, ok := metrics[nameMet]; ok {
							var ii int
							ii, err = strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
							if err != nil {
								return nil, status.Error(codes.Internal, err.Error())
							}
							valueMetInt = int(*s[i].Delta) + ii
							err = GRPCSrv.StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
							if err != nil {
								GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
							}
							err = GRPCSrv.StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
							if err != nil {
								GRPCSrv.Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
							}
						} else {
							valueMetInt = int(*s[i].Delta)
							err = GRPCSrv.StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
							if err != nil {
								GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
							}
							err = GRPCSrv.StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
							if err != nil {
								GRPCSrv.Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
							}
						}
					}
				}
			}

		} else {
			if typeMet == "gauge" {
				if s[i].Value != nil {
					if metrics[nameMet] != gauge(*s[i].Value) {
						err = GRPCSrv.StorageM.ChangeMetric(nameMet, gauge(*s[i].Value), config.ArgsM)
						if err != nil {
							GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
						}
						err = GRPCSrv.StorageM.ChangeMetricDB(nameMet, *s[i].Value, typeMet, config.ArgsM)
						if err != nil {
							GRPCSrv.Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
						}
					}
				}
			}
			//update counter
			if typeMet == "counter" {
				var valueMetInt int
				if s[i].Delta != nil {
					if _, ok := metrics[nameMet]; ok {
						var ii int
						ii, err = strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
						if err != nil {
							return nil, status.Error(codes.Internal, err.Error())
						}
						valueMetInt = int(*s[i].Delta) + ii
						err = GRPCSrv.StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
						if err != nil {
							GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
							return nil, status.Error(codes.Internal, err.Error())
						}
						err = GRPCSrv.StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
						if err != nil {
							GRPCSrv.Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
							return nil, status.Error(codes.Internal, err.Error())
						}
					} else {
						valueMetInt = int(*s[i].Delta)
						err = GRPCSrv.StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
						if err != nil {
							GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
							return nil, status.Error(codes.Internal, err.Error())
						}
						err = GRPCSrv.StorageM.ChangeMetricDB(nameMet, valueMetInt, typeMet, config.ArgsM)
						if err != nil {
							GRPCSrv.Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
							return nil, status.Error(codes.Internal, err.Error())
						}
					}
				}
			}
		}
	}
	return &pb.Empty{}, nil
}

// Get all metrics url /
func (s *grpcServer) GetMetricsJSON(ctx context.Context, empty *pb.Empty) (response *pb.GetMetricsJSONResponse, err error) {
	response = new(pb.GetMetricsJSONResponse)

	metrics, err := GRPCSrv.StorageM.GetMetricsJSON()
	if err != nil {
		GRPCSrv.Logger.Error("Error get JSON metrics", zap.Error(err))
	}

	grpcMetrics := make([]*pb.JSONMetrics, 0)
	for i := 0; i < len(metrics); i++ {
		grpcMetrics = append(grpcMetrics, converToGRPC(metrics[i]))
	}
	response.JSONMetrics = grpcMetrics
	return response, err
}

// Convert to GRPC metric
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

func (s *grpcServer) UpdateMetric(ctx context.Context, m *pb.MetricData) (p *pb.Empty, err error) {
	p = &pb.Empty{}
	typeMet := m.Mtype
	nameMet := m.ID
	value := m.Data
	// Get metrics from memory
	metrics := GRPCSrv.StorageM.GetMetrics()

	switch typeMet {
	case "gauge":
		typeMet = "gauge"
	case "counter":
		typeMet = "counter"
	}
	if typeMet != "gauge" && typeMet != "counter" {
		GRPCSrv.Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Update gauge
	if typeMet == "gauge" && nameMet != "PollCount" {
		valueMetFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			GRPCSrv.Logger.Error("Error changing metric ChangeMetricDB: ", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		} else {
			if metrics[nameMet] != gauge(valueMetFloat) {
				err = GRPCSrv.StorageM.ChangeMetric(nameMet, gauge(valueMetFloat), config.ArgsM)
				if err != nil {
					GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					return nil, status.Error(codes.Internal, err.Error())

				}
			}
		}
	}
	// Update counter
	if typeMet == "counter" {
		valueMetInt, err := strconv.Atoi(value)
		if err != nil {
			GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if err == nil {
			if _, ok := metrics[nameMet]; ok {
				var i int
				i, err = strconv.Atoi(fmt.Sprintf("%v", metrics[nameMet]))
				if err != nil {
					GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					return nil, status.Error(codes.Unknown, err.Error())
				}
				valueMetInt = valueMetInt + i
				err = GRPCSrv.StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
				if err != nil {
					GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					return nil, status.Error(codes.Internal, err.Error())
				}
			} else {
				err = GRPCSrv.StorageM.ChangeMetric(nameMet, counter(valueMetInt), config.ArgsM)
				if err != nil {
					GRPCSrv.Logger.Error("Error changing metric ChangeMetric: ", zap.Error(err))
					return nil, status.Error(codes.Internal, err.Error())
				}
			}
		}
	}
	return
}

func (s *grpcServer) GetMetricData(ctx context.Context, m *pb.Metric) (result *pb.MetricData, err error) {

	metrics := GRPCSrv.StorageM.GetMetrics()

	typeMet := m.MType
	nameMet := m.ID

	switch typeMet {
	case "gauge":
		typeMet = "gauge"
	case "counter":
		typeMet = "counter"
	}
	if typeMet != "gauge" && typeMet != "counter" {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if typeMet == "gauge" && nameMet == "PollCount" {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if _, ok := metrics[nameMet]; ok {
		if typeMet == "counter" {
			i := metrics[nameMet]
			result = &pb.MetricData{
				Delta: *(*int64)(unsafe.Pointer(&i)),
			}
			return result, status.Error(codes.Internal, err.Error())
		}
	} else {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if typeMet == "gauge" && nameMet != "PollCount" {
		a := metrics[nameMet]
		result = &pb.MetricData{
			Value: *(*float64)(unsafe.Pointer(&a)),
		}
		if err != nil {
			GRPCSrv.Logger.Error("Error write bytes to req: ", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())

		}
	}
	return
}

// Convert JSON metrics
func convertToSJSONMetrics(j []*pb.JSONMetrics) (result []storage.JSONMetrics) {
	for i := 0; i < len(j); i++ {
		js := storage.JSONMetrics{
			ID:    j[i].ID,
			Hash:  j[i].Hash,
			MType: j[i].MType,
			Delta: &j[i].Delta,
			Value: &j[i].Value,
		}
		result = append(result, js)
	}
	return result
}
