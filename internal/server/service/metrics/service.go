package metrics

import (
	"context"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
)

type Service struct {
	storage metricStorage
	mapper  metricMapper
	retrier *server.Retrier
}

func New(s metricStorage, m metricMapper, retrier *server.Retrier) Service {
	return Service{storage: s, mapper: m, retrier: retrier}
}

func (s Service) SaveMetrics(ctx context.Context, requests []common.MetricRequestDto) ([]common.MetricResponseDto, error) {
	var savedMetrics []common.MetricResponseDto
	var metrics []server.Metric

	for _, request := range requests {
		metric, err := s.mapper.MapRequestToDomainModel(request)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}

	err := s.retrier.RunSQL(func() error {
		m, err := s.storage.SaveMetrics(ctx, metrics)
		if err != nil {
			return err
		}

		for _, metric := range m {
			savedMetrics = append(savedMetrics, s.mapper.MapDomainModelToResponse(metric))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return savedMetrics, nil
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
