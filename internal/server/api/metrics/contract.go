package metrics

import "github.com/desepticon55/metrics-collector/internal/common"

type metricsService interface {
	SaveMetric(request common.MetricRequestDto) (common.MetricResponseDto, error)

	FindOneMetric(metricName string, metricType common.MetricType) (common.MetricResponseDto, error)

	FindAllMetrics() []common.MetricResponseDto
}
