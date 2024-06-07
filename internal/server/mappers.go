package server

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/go-playground/validator/v10"
	"strconv"
	"strings"
)

func MapMetricRequestDtoToMetricDomainModel(dto common.MetricRequestDto) (common.Metric, error) {
	validate := validator.New()
	validate.RegisterValidation("metricTypeCheck", metricTypeValidator)
	if err := validate.Struct(dto); err != nil {
		return common.Metric{}, fmt.Errorf("request validation failed: %s", err)
	}

	switch dto.Type {
	case common.Gauge:
		value, err := strconv.ParseFloat(strings.TrimSpace(dto.Value), 64)
		if err != nil {
			return common.Metric{}, fmt.Errorf("bad request. Gauge type has incorrect value = %s. Expected float64", dto.Value)
		}
		return common.Metric{Name: dto.Name, Type: common.Gauge, Value: value}, nil
	case common.Counter:
		value, err := strconv.ParseInt(strings.TrimSpace(dto.Value), 10, 64)
		if err != nil {
			return common.Metric{}, fmt.Errorf("bad request. Counter type has incorrect value = %s. Expected int64", dto.Value)
		}
		return common.Metric{Name: dto.Name, Type: common.Counter, Value: value}, nil
	default:
		return common.Metric{}, fmt.Errorf("unsupported metric type: %s", dto.Type)
	}
}
