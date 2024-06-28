package server

//import (
//	"github.com/desepticon55/metrics-collector/internal/common"
//	"github.com/stretchr/testify/assert"
//	"testing"
//)
//
//func TestMapMetricDtoToDomainModel(t *testing.T) {
//	type expected struct {
//		result           server.Metric
//		isNeedCheckError bool
//	}
//
//	tests := []struct {
//		name     string
//		dto      common.MetricRequestDto
//		expected expected
//	}{
//		{
//			name: "Valid gauge metric",
//			dto: common.MetricRequestDto{
//				ID:  "temperature",
//				Type:  common.Gauge,
//				Value: "23.5",
//			},
//			expected: expected{
//				result: server.Metric{
//					ID:  "temperature",
//					Type:  common.Gauge,
//					Value: 23.5,
//				},
//				isNeedCheckError: false,
//			},
//		},
//		{
//			name: "Valid counter metric",
//			dto: common.MetricRequestDto{
//				ID:  "requests",
//				Type:  common.Counter,
//				Value: "100",
//			},
//			expected: expected{
//				result: server.Metric{
//					ID:  "requests",
//					Type:  common.Counter,
//					Value: int64(100),
//				},
//				isNeedCheckError: false,
//			},
//		},
//		{
//			name: "Invalid gauge value",
//			dto: common.MetricRequestDto{
//				ID:  "temperature",
//				Type:  common.Gauge,
//				Value: "invalid",
//			},
//			expected: expected{
//				result:           server.Metric{},
//				isNeedCheckError: true,
//			},
//		},
//		{
//			name: "Invalid counter value",
//			dto: common.MetricRequestDto{
//				ID:  "requests",
//				Type:  common.Counter,
//				Value: "invalid",
//			},
//			expected: expected{
//				result:           server.Metric{},
//				isNeedCheckError: true,
//			},
//		},
//		{
//			name: "Unsupported metric type",
//			dto: common.MetricRequestDto{
//				ID:  "unknown",
//				Type:  "unknown",
//				Value: "123",
//			},
//			expected: expected{
//				result:           server.Metric{},
//				isNeedCheckError: true,
//			},
//		},
//		{
//			name: "Missing name",
//			dto: common.MetricRequestDto{
//				Type:  common.Gauge,
//				Value: "23.5",
//			},
//			expected: expected{
//				result:           server.Metric{},
//				isNeedCheckError: true,
//			},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			metric, err := MapMetricRequestDtoToMetricDomainModel(test.dto)
//			if test.expected.isNeedCheckError {
//				assert.Error(t, err)
//			} else {
//				assert.NoError(t, err)
//				assert.Equal(t, test.expected.result, metric)
//			}
//		})
//	}
//}
