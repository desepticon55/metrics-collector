package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	metrics2 "github.com/desepticon55/metrics-collector/internal/server/api/metrics"
	grpc "github.com/desepticon55/metrics-collector/proto/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

type MetricsServer struct {
	grpc.UnimplementedMetricsServiceServer
	Service metrics2.MetricsService
	Config  server.Config
	Logger  *zap.Logger
}

func (s *MetricsServer) SendMetrics(ctx context.Context, req *grpc.MetricsRequest) (*grpc.MetricsResponse, error) {
	if s.Config.HashKey != "" {
		hash := sha256.Sum256(append([]byte(req.Ip), []byte(s.Config.HashKey)...))
		hashStr := hex.EncodeToString(hash[:])
		if req.Hash != hashStr {
			s.Logger.Error("Invalid HashSHA256", zap.String("header hash", req.Hash), zap.String("calculated hash", hashStr))
			return nil, status.Error(codes.InvalidArgument, "Invalid HashSHA256")
		}
	}

	if len(s.Config.TrustedSubnet) != 0 {
		agentIP := req.Ip
		if agentIP == "" {
			s.Logger.Error("X-Real-IP header missing", zap.String("agent ip", req.Ip))
			return nil, status.Error(codes.InvalidArgument, "X-Real-IP header missing")
		}

		if !isIPInTrustedSubnet(agentIP, s.Config.TrustedSubnet) {
			s.Logger.Error("Forbidden: IP not in trusted subnet", zap.String("agent ip", req.Ip))
			return nil, status.Error(codes.InvalidArgument, "Forbidden: IP not in trusted subnet")
		}
	}

	var metrics []common.MetricRequestDto
	for _, metric := range req.Metrics {
		metrics = append(metrics, common.MetricRequestDto{
			ID:    metric.Id,
			MType: common.MetricType(metric.Type),
			Delta: &metric.Delta,
			Value: &metric.Value,
		})
	}

	_, err := s.Service.SaveMetrics(ctx, metrics)
	if err != nil {
		s.Logger.Error("Internal server error", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	return &grpc.MetricsResponse{Status: "ok"}, nil
}

func isIPInTrustedSubnet(ipStr, subnetStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, trustedNet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return false
	}

	return trustedNet.Contains(ip)
}
