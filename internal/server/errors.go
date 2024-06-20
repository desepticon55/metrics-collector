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

type IncorrectMetricValueError struct {
	metricName   string
	metricType   common.MetricType
	value        interface{}
	expectedType string
}

func (e *IncorrectMetricValueError) Error() string {
	return fmt.Sprintf(
		"Metric with name = %s and type = %s has incorrect value = %s. Expected type = %s",
		e.metricName, e.metricType, e.value, e.expectedType,
	)
}

func NewIncorrectMetricValueError(metricName string, metricType common.MetricType, value interface{}, expectedType string) *IncorrectMetricValueError {
	return &IncorrectMetricValueError{metricName: metricName, metricType: metricType, value: value, expectedType: expectedType}
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
