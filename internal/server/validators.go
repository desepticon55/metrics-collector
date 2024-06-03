package server

import "github.com/go-playground/validator/v10"

func metricTypeValidator(fl validator.FieldLevel) bool {
	metricType := MetricType(fl.Field().String())
	if metricType == "" {
		return false
	}
	allowedMetricTypes := []MetricType{Gauge, Counter}
	for _, t := range allowedMetricTypes {
		if metricType == t {
			return true
		}
	}
	return false
}
