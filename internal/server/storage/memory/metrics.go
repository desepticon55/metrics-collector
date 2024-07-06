package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"log"
	"os"
	"sync"
	"time"
)

type Storage struct {
	mu              sync.Mutex
	metrics         map[string]server.Metric
	file            string
	autoSaveEnabled bool
	saveInterval    time.Duration
}

func New(file string, isNeedLoadData bool, saveInterval time.Duration) *Storage {
	storage := &Storage{
		metrics:         make(map[string]server.Metric),
		file:            file,
		autoSaveEnabled: saveInterval > 0,
		saveInterval:    saveInterval,
	}
	if isNeedLoadData {
		err := storage.loadFromFile()
		if err != nil {
			log.Printf("Error during load metrics from file: %v", err)
		}
	}
	if storage.autoSaveEnabled {
		storage.startAutoSave()
	}

	return storage
}

func (s *Storage) SaveMetric(ctx context.Context, metric server.Metric) (server.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%s_%s", metric.GetName(), metric.GetType())
	foundMetric, exists := s.metrics[key]
	if !exists {
		s.metrics[key] = metric
	} else {
		if metric.GetType() == common.Counter {
			foundCounter := foundMetric.(*server.Counter)
			newCounter := metric.(*server.Counter)
			foundCounter.Value += newCounter.Value
			s.metrics[key] = foundCounter
		} else {
			s.metrics[key] = metric
		}
	}
	if !s.autoSaveEnabled {
		err := s.saveToFile()
		if err != nil {
			log.Printf("Error during save metric to file: %v", err)
			return nil, err
		}
	}
	return s.metrics[key], nil
}

func (s *Storage) FindOneMetric(ctx context.Context, metricName string, metricType common.MetricType) (server.Metric, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s_%s", metricName, metricType)
	metric, exists := s.metrics[key]
	return metric, exists
}

func (s *Storage) FindAllMetrics(ctx context.Context) ([]server.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	values := make([]server.Metric, 0, len(s.metrics))
	for _, value := range s.metrics {
		values = append(values, value)
	}
	return values, nil
}

func (s *Storage) loadFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var metricsMap map[string]json.RawMessage
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metricsMap); err != nil {
		return err
	}

	for key, raw := range metricsMap {
		metric, err := server.UnmarshalMetric(raw)
		if err != nil {
			return err
		}
		s.metrics[key] = metric
	}

	return nil
}

func (s *Storage) saveToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.file)
	if err != nil {
		return err
	}
	defer file.Close()

	metricsMap := make(map[string]json.RawMessage)
	for key, metric := range s.metrics {
		data, err := server.MarshalMetric(metric)
		if err != nil {
			return err
		}
		metricsMap[key] = data
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(metricsMap)
}

func (s *Storage) startAutoSave() {
	if !s.autoSaveEnabled {
		log.Printf("Auto save is disabled")
		return
	}

	go func() {
		for range time.Tick(s.saveInterval) {
			err := s.saveToFile()
			if err != nil {
				log.Printf("Error during save metrics to file: %v", err)
			}
		}
	}()
}
