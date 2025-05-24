package server

import (
	"context"

	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/models"
	pb "github.com/sejo412/ya-metrics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	pb.UnimplementedMetricsServer
	opts config.Options
}

func NewGRPCServer() *GRPCServer {
	return &GRPCServer{}
}

func RegisterGRPCServer(gRPC *grpc.Server) {
	pb.RegisterMetricsServer(gRPC, NewGRPCServer())
}

var grpcMsgUnimplemented = "unimplemented"
var grpcMsgErr = "error"

func (g *GRPCServer) SendMetrics(ctx context.Context, in *pb.SendMetricsRequest) (*pb.SendMetricsResponse, error) {
	metrics := models.ConvertPbsToV1s(in.GetMetrics())

	if err := g.opts.Storage.MassUpsert(ctx, metrics); err != nil {
		return &pb.SendMetricsResponse{Error: &grpcMsgErr}, status.Errorf(codes.Internal, "internal error: %v", err)
	}
	return nil, nil
}

func (g *GRPCServer) GetMetric(ctx context.Context, _ *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	// TODO implement me
	return &pb.GetMetricResponse{
		Metric: nil,
		Error:  &grpcMsgUnimplemented,
	}, status.Error(codes.Unimplemented, grpcMsgUnimplemented)
}

func (g *GRPCServer) GetMetrics(ctx context.Context, _ *emptypb.Empty) (*pb.GetMetricsResponse, error) {
	// TODO implement me
	return &pb.GetMetricsResponse{
		Metrics: nil,
		Error:   &grpcMsgUnimplemented,
	}, status.Error(codes.Unimplemented, grpcMsgUnimplemented)
}

func (g *GRPCServer) PingStorage(ctx context.Context, _ *emptypb.Empty) (*pb.PingStorageResponse, error) {
	ok := new(bool)
	if err := g.opts.Storage.Ping(ctx); err != nil {
		*ok = false
		return &pb.PingStorageResponse{Ok: ok}, status.Error(codes.ResourceExhausted, err.Error())
	}
	*ok = true
	return &pb.PingStorageResponse{
		Ok: ok,
	}, nil
}
