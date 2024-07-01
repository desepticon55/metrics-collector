package memory

import (
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

func (s *Storage) SaveMetric(metric server.Metric) (server.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%s_%s", metric.Name, metric.Type)
	foundMetric, exists := s.metrics[key]
	if !exists {
		s.metrics[key] = metric
	} else {
		if metric.Type == common.Counter {
			s.metrics[key] = server.Metric{
				Name:      metric.Name,
				Type:      metric.Type,
				Value:     metric.Value.(int64) + foundMetric.Value.(int64),
				ValueType: "int64",
			}
		} else {
			s.metrics[key] = metric
		}
	}
	if !s.autoSaveEnabled {
		err := s.saveToFile()
		if err != nil {
			log.Printf("Error during save metric to file: %v", err)
			return server.Metric{}, err
		}
	}
	return s.metrics[key], nil
}

func (s *Storage) FindOneMetric(metricName string, metricType common.MetricType) (server.Metric, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s_%s", metricName, metricType)
	metric, exists := s.metrics[key]
	return metric, exists
}

func (s *Storage) FindAllMetrics() []server.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()

	values := make([]server.Metric, 0, len(s.metrics))
	for _, value := range s.metrics {
		values = append(values, value)
	}
	return values
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

	decoder := json.NewDecoder(file)
	return decoder.Decode(&s.metrics)
}

func (s *Storage) saveToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.file)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(s.metrics)
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
