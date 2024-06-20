package metrics

import "github.com/desepticon55/metrics-collector/internal/common"

type metricsService interface {
	SaveMetric(request common.MetricRequestDto) error

	FindOneMetric(metricName string, metricType common.MetricType) (common.Metric, error)

	FindAllMetrics() []common.Metric
}
