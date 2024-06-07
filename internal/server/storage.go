package server

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"sync"
)

type Storage interface {
	SaveMetric(metric common.Metric) error

	GetMetric(name string, metricType common.MetricType) (common.Metric, bool)

	GetAllMetrics() []common.Metric
}

type MemStorage struct {
	mu      sync.Mutex
	metrics map[string]common.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]common.Metric),
	}
}

func (s *MemStorage) SaveMetric(metric common.Metric) error {
	s.mu.Lock()
	key := fmt.Sprintf("%s_%s", metric.Name, metric.Type)
	foundMetric, exists := s.metrics[key]
	if !exists {
		s.metrics[key] = metric
	} else {
		if metric.Type == common.Counter {
			s.metrics[key] = common.Metric{
				Name:  metric.Name,
				Type:  metric.Type,
				Value: metric.Value.(int64) + foundMetric.Value.(int64),
			}
		} else {
			s.metrics[key] = metric
		}
	}
	defer s.mu.Unlock()
	return nil
}

func (s *MemStorage) GetMetric(metricName string, metricType common.MetricType) (common.Metric, bool) {
	s.mu.Lock()
	metric, exists := s.metrics[fmt.Sprintf("%s_%s", metricName, metricType)]
	defer s.mu.Unlock()
	return metric, exists
}

func (s *MemStorage) GetAllMetrics() []common.Metric {
	s.mu.Lock()
	values := make([]common.Metric, 0, len(s.metrics))
	for _, value := range s.metrics {
		values = append(values, value)
	}
	defer s.mu.Unlock()
	return values
}
