package metrics

import "github.com/desepticon55/metrics-collector/internal/common"

type metricStorage interface {
	SaveMetric(metric common.Metric) error

	FindOneMetric(metricName string, metricType common.MetricType) (common.Metric, bool)

	FindAllMetrics() []common.Metric
}

type metricMapper interface {
	MapRequestToDomainModel(request common.MetricRequestDto) (common.Metric, error)
}
