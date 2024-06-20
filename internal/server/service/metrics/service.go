package metrics

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
)

type Service struct {
	storage metricStorage
	mapper  metricMapper
}

func New(s metricStorage, m metricMapper) Service {
	return Service{storage: s, mapper: m}
}

func (s Service) SaveMetric(request common.MetricRequestDto) error {
	metric, err := s.mapper.MapRequestToDomainModel(request)
	if err != nil {
		return err
	}

	return s.storage.SaveMetric(metric)
}

func (s Service) FindOneMetric(metricName string, metricType common.MetricType) (common.Metric, error) {
	metric, exist := s.storage.FindOneMetric(metricName, metricType)
	if !exist {
		return common.Metric{}, server.NewMetricNotFoundError(metricName, metricType)
	}
	return metric, nil
}

func (s Service) FindAllMetrics() []common.Metric {
	return s.storage.FindAllMetrics()
}
