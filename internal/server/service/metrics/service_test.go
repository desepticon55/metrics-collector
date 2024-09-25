package metrics

import (
	"context"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockMetricStorage struct {
	mock.Mock
}

func (m *mockMetricStorage) SaveMetrics(ctx context.Context, metrics []server.Metric) ([]server.Metric, error) {
	args := m.Called(ctx, metrics)
	return args.Get(0).([]server.Metric), args.Error(1)
}

func (m *mockMetricStorage) FindOneMetric(ctx context.Context, name string, mType common.MetricType) (server.Metric, bool) {
	args := m.Called(ctx, name, mType)
	return args.Get(0).(server.Metric), args.Bool(1)
}

func (m *mockMetricStorage) FindAllMetrics(ctx context.Context) ([]server.Metric, error) {
	args := m.Called(ctx)
	return args.Get(0).([]server.Metric), args.Error(1)
}

type mockMetricMapper struct {
	mock.Mock
}

func (m *mockMetricMapper) MapRequestToDomainModel(request common.MetricRequestDto) (server.Metric, error) {
	args := m.Called(request)
	return args.Get(0).(server.Metric), args.Error(1)
}

func (m *mockMetricMapper) MapDomainModelToResponse(metric server.Metric) common.MetricResponseDto {
	args := m.Called(metric)
	return args.Get(0).(common.MetricResponseDto)
}

type mockRetrier struct {
	mock.Mock
}

func (r *mockRetrier) RunSQL(fn func() error) error {
	args := r.Called(fn)
	return args.Error(0)
}

func TestService_FindOneMetric(t *testing.T) {
	ctx := context.Background()

	storage := new(mockMetricStorage)
	mapper := new(mockMetricMapper)

	service := New(storage, mapper, nil)

	metricName := "test_metric"
	metricType := common.Gauge

	foundMetric := &server.Gauge{BaseMetric: server.BaseMetric{Name: "metric2", Type: common.Gauge}, Value: 99.9}
	storage.On("FindOneMetric", ctx, metricName, metricType).Return(foundMetric, true)

	response := common.MetricResponseDto{ID: metricName, MType: common.Gauge, Value: new(float64)}
	mapper.On("MapDomainModelToResponse", foundMetric).Return(response)

	metric, err := service.FindOneMetric(ctx, metricName, metricType)
	assert.NoError(t, err)
	assert.Equal(t, response, metric)

	storage.AssertExpectations(t)
	mapper.AssertExpectations(t)
}

func TestService_FindAllMetrics(t *testing.T) {
	ctx := context.Background()

	storage := new(mockMetricStorage)
	mapper := new(mockMetricMapper)

	service := New(storage, mapper, nil)

	metrics := []server.Metric{
		&server.Counter{BaseMetric: server.BaseMetric{Name: "counter_metric", Type: common.Counter}, Value: 100},
		&server.Gauge{BaseMetric: server.BaseMetric{Name: "gauge_metric", Type: common.Gauge}, Value: 50.5},
	}
	storage.On("FindAllMetrics", ctx).Return(metrics, nil)

	response1 := common.MetricResponseDto{ID: "counter_metric", MType: common.Counter, Delta: new(int64)}
	response2 := common.MetricResponseDto{ID: "gauge_metric", MType: common.Gauge, Value: new(float64)}
	mapper.On("MapDomainModelToResponse", metrics[0]).Return(response1)
	mapper.On("MapDomainModelToResponse", metrics[1]).Return(response2)

	allMetrics := service.FindAllMetrics(ctx)
	assert.Len(t, allMetrics, 2)
	assert.Equal(t, response1, allMetrics[0])
	assert.Equal(t, response2, allMetrics[1])

	storage.AssertExpectations(t)
	mapper.AssertExpectations(t)
}
