package metrics

import (
	"context"
	"github.com/desepticon55/metrics-collector/internal/common"
)

type MetricsService interface {
	SaveMetrics(ctx context.Context, request []common.MetricRequestDto) ([]common.MetricResponseDto, error)

	FindOneMetric(ctx context.Context, metricName string, metricType common.MetricType) (common.MetricResponseDto, error)

	FindAllMetrics(ctx context.Context) []common.MetricResponseDto
}
