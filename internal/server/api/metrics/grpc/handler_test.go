package grpc

import (
	"context"
	"github.com/desepticon55/metrics-collector/internal/common"
	server2 "github.com/desepticon55/metrics-collector/internal/server"
	grpc "github.com/desepticon55/metrics-collector/proto/metrics"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) SaveMetrics(ctx context.Context, request []common.MetricRequestDto) ([]common.MetricResponseDto, error) {
	args := m.Called(ctx, request)
	return args.Get(0).([]common.MetricResponseDto), args.Error(1)
}

func (m *MockMetricsService) FindOneMetric(ctx context.Context, metricName string, metricType common.MetricType) (common.MetricResponseDto, error) {
	panic("implement me")
}

func (m *MockMetricsService) FindAllMetrics(ctx context.Context) []common.MetricResponseDto {
	panic("implement me")
}

func TestSendMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mockService := new(MockMetricsService)
	mockService.On("SaveMetrics", mock.Anything, mock.Anything).Return([]common.MetricResponseDto{}, nil)

	tests := []struct {
		name        string
		config      server2.Config
		req         *grpc.MetricsRequest
		expectedErr error
	}{
		{
			name: "valid request",
			config: server2.Config{
				HashKey:       "somehashkey",
				TrustedSubnet: "192.168.1.0/24",
			},
			req: &grpc.MetricsRequest{
				Ip:   "192.168.1.2",
				Hash: "d55a6e05957298d3c93e6bf8b48e903e55c3c7a9a4b4d20e2d39259b0de4ce3e",
				Metrics: []*grpc.Metric{
					{Id: "metric1", Type: "gauge", Delta: 10, Value: 5.5},
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid hash",
			config: server2.Config{
				HashKey:       "somehashkey",
				TrustedSubnet: "192.168.1.0/24",
			},
			req: &grpc.MetricsRequest{
				Ip:   "192.168.1.2",
				Hash: "invalidhash",
				Metrics: []*grpc.Metric{
					{Id: "metric1", Type: "gauge", Delta: 10, Value: 5.5},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "Invalid HashSHA256"),
		},
		{
			name: "missing IP",
			config: server2.Config{
				TrustedSubnet: "192.168.1.0/24",
			},
			req: &grpc.MetricsRequest{
				Ip:   "",
				Hash: "d55a6e05957298d3c93e6bf8b48e903e55c3c7a9a4b4d20e2d39259b0de4ce3e",
				Metrics: []*grpc.Metric{
					{Id: "metric1", Type: "gauge", Delta: 10, Value: 5.5},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "X-Real-IP header missing"),
		},
		{
			name: "IP not in trusted subnet",
			config: server2.Config{
				TrustedSubnet: "192.168.1.0/24",
			},
			req: &grpc.MetricsRequest{
				Ip:   "10.0.0.1",
				Hash: "d55a6e05957298d3c93e6bf8b48e903e55c3c7a9a4b4d20e2d39259b0de4ce3e",
				Metrics: []*grpc.Metric{
					{Id: "metric1", Type: "gauge", Delta: 10, Value: 5.5},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "Forbidden: IP not in trusted subnet"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &MetricsServer{
				Config:  tt.config,
				Logger:  logger,
				Service: mockService,
			}

			_, err := server.SendMetrics(context.Background(), tt.req)

			if tt.expectedErr != nil {
				actualErrCode := status.Code(err)
				expectedErrCode := status.Code(tt.expectedErr)

				assert.Equal(t, expectedErrCode, actualErrCode)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
