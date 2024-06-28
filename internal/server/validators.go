package server

import (
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/go-playground/validator/v10"
	"slices"
)

func MetricValidator(sl validator.StructLevel) {
	dto := sl.Current().Interface().(common.MetricRequestDto)
	allowedMetricTypes := []common.MetricType{common.Gauge, common.Counter}

	if !slices.Contains(allowedMetricTypes, dto.MType) {
		sl.ReportError(dto.MType, "MType", "type", "supported", "")
	}
	if dto.MType == common.Counter && dto.Delta == nil {
		sl.ReportError(dto.Delta, "Delta", "delta", "required", "")
	}
	if dto.MType == common.Gauge && dto.Value == nil {
		sl.ReportError(dto.Value, "Value", "value", "required", "")
	}
}
