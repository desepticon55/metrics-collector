package memory

import (
	"context"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestStorage_AutoSave(t *testing.T) {
	file, err := os.CreateTemp("", "metrics_storage_auto_save_test_*.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	storage := New(file.Name(), false, 100*time.Millisecond)

	counterMetric := &server.Counter{BaseMetric: server.BaseMetric{Name: "requests", Type: common.Counter}, Value: 10}
	metrics := []server.Metric{counterMetric}

	_, err = storage.SaveMetrics(context.Background(), metrics)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	loadedStorage := New(file.Name(), true, 0)
	foundMetric, exists := loadedStorage.FindOneMetric(context.Background(), "requests", common.Counter)
	assert.True(t, exists)
	assert.Equal(t, counterMetric, foundMetric)
}
