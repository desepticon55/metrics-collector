package metrics

import (
	"context"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
)

type metricStorage interface {
	SaveMetrics(ctx context.Context, metrics []server.Metric) ([]server.Metric, error)

	FindOneMetric(ctx context.Context, metricName string, metricType common.MetricType) (server.Metric, bool)

	FindAllMetrics(ctx context.Context) ([]server.Metric, error)
}

type metricMapper interface {
	MapRequestToDomainModel(request common.MetricRequestDto) (server.Metric, error)

	MapDomainModelToResponse(domainModel server.Metric) common.MetricResponseDto
}
