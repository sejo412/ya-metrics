package server

import (
	"context"
	"encoding/json"
	"net"
	"net/netip"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/pkg/utils"
	pb "github.com/sejo412/ya-metrics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

func NewGRPCServerWithOptions(opts *config.Options) *GRPCServer {
	router := NewGRPCServer()
	router.opts.Config = opts.Config
	router.opts.Storage = opts.Storage
	router.opts.PrivateKey = opts.PrivateKey
	if opts.TrustedSubnets != nil {
		router.opts.TrustedSubnets = opts.TrustedSubnets
	} else {
		router.opts.TrustedSubnets = []net.IPNet{}
	}
	router.opts.Logger = opts.Logger
	return router
}

func RegisterGRPCServer(gRPC *grpc.Server, server *GRPCServer) {
	pb.RegisterMetricsServer(gRPC, server)
}

var grpcMsgErr = "error"

func (g *GRPCServer) SendMetrics(ctx context.Context, in *pb.SendMetricsRequest) (*pb.SendMetricsResponse, error) {
	metrics := models.ConvertPbsToV1s(in.GetMetrics())

	if err := g.opts.Storage.MassUpsert(ctx, metrics); err != nil {
		return &pb.SendMetricsResponse{Error: &grpcMsgErr}, status.Errorf(codes.Internal, "internal error: %v", err)
	}
	return nil, nil
}

func (g *GRPCServer) GetMetric(ctx context.Context, in *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	log := g.opts.Logger.Logger
	mNamePb := in.GetId()
	mTypePb := in.GetKind().String()
	metric, err := g.opts.Storage.Get(ctx, mTypePb, mNamePb)
	if err != nil {
		log.Errorw("get metric", "type", mTypePb, "id", mNamePb, "err", err)
		return &pb.GetMetricResponse{Error: &grpcMsgErr}, status.Errorf(codes.NotFound, "%v", err)
	}
	res, err := models.ConvertV1ToPb(metric)
	if err != nil {
		log.Errorw("convert metric", "type", mTypePb, "id", mNamePb, "err", err)
		return &pb.GetMetricResponse{Error: &grpcMsgErr}, status.Errorf(codes.NotFound, "%v", err)
	}
	return &pb.GetMetricResponse{
		Metric: res,
		Error:  nil,
	}, nil
}

func (g *GRPCServer) GetMetrics(ctx context.Context, _ *emptypb.Empty) (*pb.GetMetricsResponse, error) {
	log := g.opts.Logger.Logger
	m, err := g.opts.Storage.GetAll(ctx)
	if err != nil {
		log.Errorw("get metrics", "error", err)
		return &pb.GetMetricsResponse{Error: &grpcMsgErr}, status.Errorf(codes.Internal, grpcMsgErr)
	}
	metrics, err := models.ConvertV1sToPbs(m)
	if err != nil {
		log.Errorw("convert metrics", "error", err)
		return &pb.GetMetricsResponse{Error: &grpcMsgErr}, status.Errorf(codes.Internal, grpcMsgErr)
	}
	return &pb.GetMetricsResponse{
		Metrics: metrics,
		Error:   nil,
	}, nil
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

func interceptorLogger(l *logger.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, keyvals ...any) {
		l.Logger.Log(l.IntToLevel(int(lvl)), msg, keyvals)
	})
}

func interceptorXRealIPOptions(addrs []net.IPNet) []realip.Option {
	res := make([]realip.Option, 0)
	if addrs == nil {
		return res
	}
	prefixes := make([]netip.Prefix, 0)
	for _, addr := range addrs {
		prefix := netip.MustParsePrefix(addr.String())
		prefixes = append(prefixes, prefix)
	}

	res = append(res, realip.WithTrustedPeers(prefixes), realip.WithHeaders([]string{realip.XRealIp}))
	return res
}

func (g *GRPCServer) interceptorCheckHash(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	// skip if key not specified
	key := g.opts.Config.Key
	if key == "" {
		return handler(ctx, req)
	}
	var hash string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(models.HTTPHeaderSign)
		if len(values) > 0 {
			hash = values[0]
		}
	}
	if len(hash) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing hash")
	}
	r, ok := req.(pb.SendMetricsRequest)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid request")
	}
	metrics := r.GetMetrics()
	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	want := utils.Hash(metricsBytes, key)
	if want != hash {
		return nil, status.Error(codes.Unauthenticated, "invalid hash")
	}
	return handler(ctx, req)
}

func gRPCServerOptions(server *GRPCServer, key string) []grpc.ServerOption {
	res := make([]grpc.ServerOption, 0)
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0)
	unaryInterceptors = append(unaryInterceptors,
		logging.UnaryServerInterceptor(interceptorLogger(&server.opts.Logger)))
	realIpOpts := interceptorXRealIPOptions(server.opts.TrustedSubnets)
	if realIpOpts != nil {
		unaryInterceptors = append(unaryInterceptors, realip.UnaryServerInterceptorOpts(realIpOpts...))
	}
	if key != "" {
		unaryInterceptors = append(unaryInterceptors, server.interceptorCheckHash)
	}
	res = append(res, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	return res
}
