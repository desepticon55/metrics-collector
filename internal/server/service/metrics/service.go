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

func (s Service) SaveMetric(request common.MetricRequestDto) (common.MetricResponseDto, error) {
	metric, err := s.mapper.MapRequestToDomainModel(request)
	if err != nil {
		return common.MetricResponseDto{}, err
	}

	savedMetric, err := s.storage.SaveMetric(metric)
	if err != nil {
		return common.MetricResponseDto{}, err
	}
	return s.mapper.MapDomainModelToResponse(savedMetric), nil
}

func (s Service) FindOneMetric(metricName string, metricType common.MetricType) (common.MetricResponseDto, error) {
	metric, exist := s.storage.FindOneMetric(metricName, metricType)
	if !exist {
		return common.MetricResponseDto{}, server.NewMetricNotFoundError(metricName, metricType)
	}
	return s.mapper.MapDomainModelToResponse(metric), nil
}

func (s Service) FindAllMetrics() []common.MetricResponseDto {
	metrics := s.storage.FindAllMetrics()
	dtoList := make([]common.MetricResponseDto, 0, len(metrics))
	for idx, metric := range metrics {
		dtoList[idx] = s.mapper.MapDomainModelToResponse(metric)
	}
	return dtoList
}
