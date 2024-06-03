package server

type Metric struct {
	Name  string
	Type  MetricType
	Value interface{}
}

type MemStorage struct {
	metrics map[string]Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]Metric),
	}
}

func (s *MemStorage) SaveMetric(metric Metric) {
	s.metrics[metric.Name] = metric
}

func (s *MemStorage) GetMetric(name string) (Metric, bool) {
	metric, exists := s.metrics[name]
	return metric, exists
}
