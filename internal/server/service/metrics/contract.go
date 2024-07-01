package metrics

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
)

type metricStorage interface {
	SaveMetric(metric server.Metric) (server.Metric, error)

	FindOneMetric(metricName string, metricType common.MetricType) (server.Metric, bool)

	FindAllMetrics() []server.Metric
}

type metricMapper interface {
	MapRequestToDomainModel(request common.MetricRequestDto) (server.Metric, error)

	MapDomainModelToResponse(domainModel server.Metric) common.MetricResponseDto
}
