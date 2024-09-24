package agent

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRuntimeMetricProvider_GetMetrics(t *testing.T) {
	provider := &RuntimeMetricProvider{}
	metrics := provider.GetMetrics()

	assert.NotEmpty(t, metrics, "Metrics should be not empty")
	assert.Equal(t, 29, len(metrics), "Expected 29 metrics")
}

func TestVirtualMetricProvider_GetMetrics(t *testing.T) {
	provider := &VirtualMetricProvider{}
	metrics := provider.GetMetrics()

	assert.NotEmpty(t, metrics, "Metrics should be not empty")
	assert.GreaterOrEqual(t, len(metrics), 2, "Expected 2 metrics")
}

func TestMakeGaugeMetricRequest(t *testing.T) {
	value := 123.45
	metric := makeGaugeMetricRequest("TestMetric", value)

	assert.Equal(t, "TestMetric", metric.ID)
	assert.Equal(t, common.Gauge, metric.MType)
	assert.Equal(t, value, *metric.Value)
	assert.Nil(t, metric.Delta)
}

func TestMakeCounterMetricRequest(t *testing.T) {
	delta := int64(10)
	metric := makeCounterMetricRequest("TestCounter", delta)

	assert.Equal(t, "TestCounter", metric.ID)
	assert.Equal(t, common.Counter, metric.MType)
	assert.Equal(t, delta, *metric.Delta)
	assert.Nil(t, metric.Value)
}
