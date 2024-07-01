package metrics

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapRequestToDomainModel(t *testing.T) {
	mapper := NewMapper()

	tests := []struct {
		name      string
		dto       common.MetricRequestDto
		expected  server.Metric
		expectErr bool
	}{
		{
			name: "Gauge metric",
			dto: common.MetricRequestDto{
				ID:    "test_gauge",
				MType: "gauge",
				Value: newFloat64(123.45),
			},
			expected: server.Metric{
				Name:      "test_gauge",
				Type:      common.Gauge,
				Value:     123.45,
				ValueType: "float64",
			},
			expectErr: false,
		},
		{
			name: "Counter metric",
			dto: common.MetricRequestDto{
				ID:    "test_counter",
				MType: "counter",
				Delta: newInt64(100),
			},
			expected: server.Metric{
				Name:      "test_counter",
				Type:      common.Counter,
				Value:     int64(100),
				ValueType: "int64",
			},
			expectErr: false,
		},
		{
			name: "Unsupported metric type",
			dto: common.MetricRequestDto{
				ID:    "test_unknown",
				MType: "unknown",
			},
			expected:  server.Metric{},
			expectErr: true,
		},
		{
			name: "Invalid Gauge metric without value",
			dto: common.MetricRequestDto{
				ID:    "test_invalid_gauge",
				MType: "gauge",
			},
			expected:  server.Metric{},
			expectErr: true,
		},
		{
			name: "Invalid Counter metric without delta",
			dto: common.MetricRequestDto{
				ID:    "test_invalid_counter",
				MType: "counter",
			},
			expected:  server.Metric{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper.MapRequestToDomainModel(tt.dto)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMapDomainModelToResponse(t *testing.T) {
	mapper := NewMapper()

	tests := []struct {
		name     string
		domain   server.Metric
		expected common.MetricResponseDto
	}{
		{
			name: "Gauge metric",
			domain: server.Metric{
				Name:      "test_gauge",
				Type:      common.Gauge,
				Value:     123.45,
				ValueType: "float64",
			},
			expected: common.MetricResponseDto{
				ID:    "test_gauge",
				MType: common.Gauge,
				Value: newFloat64(123.45),
				Delta: nil,
			},
		},
		{
			name: "Counter metric",
			domain: server.Metric{
				Name:      "test_counter",
				Type:      common.Counter,
				Value:     int64(100),
				ValueType: "int64",
			},
			expected: common.MetricResponseDto{
				ID:    "test_counter",
				MType: common.Counter,
				Value: nil,
				Delta: newInt64(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapDomainModelToResponse(tt.domain)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func newFloat64(f float64) *float64 {
	return &f
}

func newInt64(i int64) *int64 {
	return &i
}
