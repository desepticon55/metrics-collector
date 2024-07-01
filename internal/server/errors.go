package server

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
)

type MetricNotFoundError struct {
	metricName string
	metricType common.MetricType
}

func (e *MetricNotFoundError) Error() string {
	return fmt.Sprintf("Metric with name = %s and type = %s was not found", e.metricName, e.metricType)
}

func NewMetricNotFoundError(metricName string, metricType common.MetricType) *MetricNotFoundError {
	return &MetricNotFoundError{metricName: metricName, metricType: metricType}
}

type ValidationError struct {
	error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("request validation failed: %s", e.error)
}

func NewValidationError(err error) *ValidationError {
	return &ValidationError{err}
}
