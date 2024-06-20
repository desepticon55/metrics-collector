package server

//
//import (
//	"github.com/desepticon55/metrics-collector/internal/common"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"testing"
//)
//
//func TestMemStorage_SaveMetric(t *testing.T) {
//	storage := NewMemStorage()
//
//	require.NoError(t, storage.SaveMetric(common.Metric{Name: "some_metric", Type: common.Gauge, Value: 1.85}))
//
//	_, exists := storage.metrics["some_metric_gauge"]
//	assert.True(t, exists)
//}
//
//func TestMemStorage_GetMetric(t *testing.T) {
//	storage := NewMemStorage()
//	require.NoError(t, storage.SaveMetric(common.Metric{Name: "some_metric", Type: common.Gauge, Value: 1.85}))
//
//	_, exists := storage.GetMetric("some_metric", common.Gauge)
//	assert.True(t, exists)
//}
//
//func TestMemStorage_GetAllMetrics(t *testing.T) {
//	storage := NewMemStorage()
//	require.NoError(t, storage.SaveMetric(common.Metric{Name: "some_gauge_metric", Type: common.Gauge, Value: 1.85}))
//	require.NoError(t, storage.SaveMetric(common.Metric{Name: "some_counter_metric", Type: common.Counter, Value: 1}))
//
//	result := storage.GetAllMetrics()
//
//	expected := []common.Metric{
//		{Name: "some_gauge_metric", Type: common.Gauge, Value: 1.85},
//		{Name: "some_counter_metric", Type: common.Counter, Value: 1},
//	}
//	assert.ElementsMatch(t, result, expected)
//}
