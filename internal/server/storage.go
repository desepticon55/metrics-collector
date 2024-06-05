package server

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"sync"
)

type Storage interface {
	SaveMetric(metric common.Metric)

	GetMetric(name string) (common.Metric, bool)
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

func (s *MemStorage) SaveMetric(metric common.Metric) {
	s.mu.Lock()
	s.metrics[metric.Name] = metric
	s.mu.Unlock()
}

func (s *MemStorage) GetMetric(name string) (common.Metric, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	metric, exists := s.metrics[name]
	return metric, exists
}
