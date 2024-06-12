package server

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/go-playground/validator/v10"
)

func metricTypeValidator(fl validator.FieldLevel) bool {
	metricType := common.MetricType(fl.Field().String())
	if metricType == "" {
		return false
	}
	allowedMetricTypes := []common.MetricType{common.Gauge, common.Counter}
	for _, t := range allowedMetricTypes {
		if metricType == t {
			return true
		}
	}
	return false
}
