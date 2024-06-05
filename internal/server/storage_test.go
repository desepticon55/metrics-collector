package server

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_SaveMetric(t *testing.T) {
	storage := NewMemStorage()

	storage.SaveMetric(common.Metric{Name: "some_metric", Type: common.Gauge, Value: 1.85})

	_, exists := storage.metrics["some_metric"]
	assert.True(t, exists)
}

func TestMemStorage_GetMetric(t *testing.T) {
	storage := NewMemStorage()
	storage.SaveMetric(common.Metric{Name: "some_metric", Type: common.Gauge, Value: 1.85})

	_, exists := storage.GetMetric("some_metric")
	assert.True(t, exists)
}
