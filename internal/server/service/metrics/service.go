package metrics

import (
	"context"
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

func (s Service) SaveMetric(ctx context.Context, request common.MetricRequestDto) (common.MetricResponseDto, error) {
	metric, err := s.mapper.MapRequestToDomainModel(request)
	if err != nil {
		return common.MetricResponseDto{}, err
	}

	savedMetric, err := s.storage.SaveMetric(ctx, metric)
	if err != nil {
		return common.MetricResponseDto{}, err
	}
	return s.mapper.MapDomainModelToResponse(savedMetric), nil
}

func (s Service) FindOneMetric(ctx context.Context, metricName string, metricType common.MetricType) (common.MetricResponseDto, error) {
	metric, exist := s.storage.FindOneMetric(ctx, metricName, metricType)
	if !exist {
		return common.MetricResponseDto{}, server.NewMetricNotFoundError(metricName, metricType)
	}
	return s.mapper.MapDomainModelToResponse(metric), nil
}

func (s Service) FindAllMetrics(ctx context.Context) []common.MetricResponseDto {
	metrics, err := s.storage.FindAllMetrics(ctx)
	if err != nil {
		return make([]common.MetricResponseDto, 0)
	}
	dtoList := make([]common.MetricResponseDto, 0, len(metrics))
	for _, metric := range metrics {
		dtoList = append(dtoList, s.mapper.MapDomainModelToResponse(metric))
	}
	return dtoList
}
