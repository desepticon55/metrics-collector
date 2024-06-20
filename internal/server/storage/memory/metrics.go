package memory

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"sync"
)

type Storage struct {
	mu      sync.Mutex
	metrics map[string]common.Metric
}

func New() *Storage {
	return &Storage{
		metrics: make(map[string]common.Metric),
	}
}

func (s *Storage) SaveMetric(metric common.Metric) error {
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

func (s *Storage) FindOneMetric(metricName string, metricType common.MetricType) (common.Metric, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s_%s", metricName, metricType)
	metric, exists := s.metrics[key]
	return metric, exists
}

func (s *Storage) FindAllMetrics() []common.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()

	values := make([]common.Metric, 0, len(s.metrics))
	for _, value := range s.metrics {
		values = append(values, value)
	}
	return values
}
